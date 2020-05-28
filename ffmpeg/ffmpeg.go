package ffmpeg

import (
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/core"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertVideoToAudio(media *core.Media) error {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return errors.New("ffmpeg not found in path")
	} else {
		videoFileName := media.FilePath
		audioFileName := strings.TrimRight(media.FilePath, filepath.Ext(media.FilePath)) + ".mp3"
		cmd := exec.Command(ffmpeg, "-y", "-loglevel", "quiet", "-i", videoFileName, "-vn", audioFileName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to extract audio: %s", err.Error())
		}
		_ = os.Remove(videoFileName)
		media.FilePath = audioFileName
		return nil
	}
}

func ConvertMp3ToDca(media *core.Media) error {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return errors.New("ffmpeg not found in path")
	} else {
		mp3FileName := media.FilePath
		dcaFileName := media.FilePath + ".dca"
		args := []string{
			"-stats",
			"-i", mp3FileName,
			"-reconnect", "1",
			"-reconnect_at_eof", "1",
			"-reconnect_streamed", "1",
			"-reconnect_delay_max", "2",
			"-map", "0:a",
			"-acodec", "libopus",
			"-f", "ogg",
			"-vbr", "on",
			"-compression_level", "10",
			"-vol", "256",
			"-ar", "48000",
			"-ac", "2",
			"-b:a", "96000",
			"-application", "audio",
			"-frame_duration", "20",
			"-packet_loss", "1",
			"-threads", "0",
			"-y", // overwrite output
			dcaFileName,
		}
		cmd := exec.Command(ffmpeg, args...)
		log.Printf("ffmpeg %v", args)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		//_ = os.Remove(mp3FileName)
		media.FilePath = dcaFileName
		return nil
	}
}
