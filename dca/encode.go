package dca

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/jonas747/ogg"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var (
	ErrBadFrame = errors.New("bad frame")
)

// EncodeOptions is a set of options for encoding dca
type EncodeOptions struct {
	Volume           int  // change audio volume (256=normal)
	FrameRate        int  // audio sampling rate (ex 48000)
	FrameDuration    int  // audio frame duration can be 20, 40, or 60 (ms)
	Bitrate          int  // audio encoding bitrate in kb/s can be 8 - 128
	PacketLoss       int  // expected packet loss percentage
	CompressionLevel int  // Compression level, higher is better quality but slower encoding (0 - 10)
	BufferedFrames   int  // How big the frame buffer should be
	VariableBitrate  bool // Whether vbr is used or not (variable bitrate)
	Threads          int  // Number of threads to use, 0 for auto
}

// StdEncodeOptions is the standard options for encoding
var StdEncodeOptions = &EncodeOptions{
	Volume:           256,
	FrameRate:        48000,
	FrameDuration:    20,
	Bitrate:          64,
	CompressionLevel: 10,
	PacketLoss:       1,
	BufferedFrames:   100, // At 20ms frames that's 2s
	VariableBitrate:  true,
}

type Frame struct {
	data     []byte
	metaData bool
}

type EncodeSession struct {
	sync.Mutex
	options    *EncodeOptions
	pipeReader io.Reader
	filePath   string

	running      bool
	started      time.Time
	frameChannel chan *Frame
	process      *os.Process

	lastFrame int
	err       error

	ffmpegOutput string

	// buffer that stores unread bytes (not full frames)
	// used to implement io.Reader
	buf bytes.Buffer
}

// EncodeFile encodes the file/url/other in path
func EncodeFile(path string, options *EncodeOptions) (session *EncodeSession, err error) {
	session = &EncodeSession{
		options:      options,
		filePath:     path,
		frameChannel: make(chan *Frame, options.BufferedFrames),
	}
	go session.run()
	return
}

func (e *EncodeSession) run() {
	// Reset running state
	defer func() {
		e.Lock()
		e.running = false
		e.Unlock()
	}()

	e.Lock()
	e.running = true

	inputFile := "pipe:0"
	if e.filePath != "" {
		inputFile = e.filePath
	}

	if e.options == nil {
		e.options = StdEncodeOptions
	}

	args := []string{
		"-i", inputFile,
		"-reconnect", "1",
		"-reconnect_at_eof", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "2",
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "ogg",
		"-compression_level", strconv.Itoa(e.options.CompressionLevel),
		"-vol", strconv.Itoa(e.options.Volume),
		"-ar", strconv.Itoa(e.options.FrameRate),
		"-ac", "2",
		"-b:a", strconv.Itoa(e.options.Bitrate * 1000),
		"-application", "audio",
		"-frame_duration", strconv.Itoa(e.options.FrameDuration),
		"-packet_loss", strconv.Itoa(e.options.PacketLoss),
		"-threads", strconv.Itoa(e.options.Threads),
		"pipe:1",
	}

	ffmpeg := exec.Command("ffmpeg", args...)

	if e.pipeReader != nil {
		ffmpeg.Stdin = e.pipeReader
	}

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		e.Unlock()
		log.Println("StdoutPipe Error:", err)
		close(e.frameChannel)
		return
	}

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		e.Unlock()
		log.Println("StderrPipe Error:", err)
		close(e.frameChannel)
		return
	}

	// Starts the ffmpeg command
	err = ffmpeg.Start()
	if err != nil {
		e.Unlock()
		log.Println("RunStart Error:", err)
		close(e.frameChannel)
		return
	}

	e.started = time.Now()

	e.process = ffmpeg.Process
	e.Unlock()

	var wg sync.WaitGroup
	wg.Add(1)
	go e.readStderr(stderr, &wg)

	defer close(e.frameChannel)
	e.readStdout(stdout)
	wg.Wait()
	err = ffmpeg.Wait()
	if err != nil {
		if err.Error() != "signal: killed" {
			e.Lock()
			e.err = err
			e.Unlock()
		}
	}
}

func (e *EncodeSession) readStderr(stderr io.ReadCloser, wg *sync.WaitGroup) {
	defer wg.Done()

	bufReader := bufio.NewReader(stderr)
	var outBuf bytes.Buffer
	for {
		r, _, err := bufReader.ReadRune()
		if err != nil {
			if err != io.EOF {
				log.Println("Error Reading stderr:", err)
			}
			break
		}

		// Is this the best way to distinguish stats from messages?
		switch r {
		case '\n':
			// Message
			e.Lock()
			e.ffmpegOutput += outBuf.String() + "\n"
			e.Unlock()
			outBuf.Reset()
		default:
			outBuf.WriteRune(r)
		}
	}
}

func (e *EncodeSession) readStdout(stdout io.ReadCloser) {
	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(stdout))

	// the first 2 packets are ogg opus metadata
	skipPackets := 2
	for {
		// Retrieve a packet
		packet, _, err := decoder.Decode()
		if skipPackets > 0 {
			skipPackets--
			continue
		}
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading ffmpeg stdout:", err)
			}
			break
		}

		err = e.writeOpusFrame(packet)
		if err != nil {
			log.Println("Error writing opus frame:", err)
			break
		}
	}
}

func (e *EncodeSession) writeOpusFrame(opusFrame []byte) error {
	var dcaBuf bytes.Buffer
	err := binary.Write(&dcaBuf, binary.LittleEndian, int16(len(opusFrame)))
	if err != nil {
		return err
	}

	_, err = dcaBuf.Write(opusFrame)
	if err != nil {
		return err
	}

	e.frameChannel <- &Frame{dcaBuf.Bytes(), false}

	e.Lock()
	e.lastFrame++
	e.Unlock()

	return nil
}

// Stop stops the encoding session
func (e *EncodeSession) Stop() error {
	e.Lock()
	defer e.Unlock()
	return e.process.Kill()
}

// ReadFrame blocks until a frame is read or there are no more frames
// Note: If rawoutput is not set, the first frame will be a metadata frame
func (e *EncodeSession) ReadFrame() (frame []byte, err error) {
	f := <-e.frameChannel
	if f == nil {
		return nil, io.EOF
	}

	return f.data, nil
}

// OpusFrame implements OpusReader, returning the next opus frame
func (e *EncodeSession) OpusFrame() (frame []byte, err error) {
	f := <-e.frameChannel
	if f == nil {
		return nil, io.EOF
	}

	if f.metaData {
		// Return the next one then...
		return e.OpusFrame()
	}

	if len(f.data) < 2 {
		return nil, ErrBadFrame
	}

	return f.data[2:], nil
}

// Running returns true if running
func (e *EncodeSession) Running() (running bool) {
	e.Lock()
	running = e.running
	e.Unlock()
	return
}

// Truncate is deprecated, use Cleanup instead
// this will be removed in a future version
func (e *EncodeSession) Truncate() {
	e.Cleanup()
}

// Cleanup cleans up the encoding session, throwing away all unread frames and stopping ffmpeg
// ensuring that no ffmpeg processes starts piling up on your system
// You should always call this after it's done
func (e *EncodeSession) Cleanup() {
	_ = e.Stop()

	for range e.frameChannel {
		// empty till closed
	}
}

// Read implements io.Reader,
// n == len(p) if err == nil, otherwise n contains the number bytes read before an error occured
func (e *EncodeSession) Read(p []byte) (n int, err error) {
	if e.buf.Len() >= len(p) {
		return e.buf.Read(p)
	}
	for e.buf.Len() < len(p) {
		f, err := e.ReadFrame()
		if err != nil {
			break
		}
		e.buf.Write(f)
	}
	return e.buf.Read(p)
}

// FrameDuration implements OpusReader, retruning the duratio of each frame
func (e *EncodeSession) FrameDuration() time.Duration {
	return time.Duration(e.options.FrameDuration) * time.Millisecond
}
