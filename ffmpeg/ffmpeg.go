package ffmpeg

import (
	"errors"
	"fmt"
	"github.com/TwinProduction/discord-music-bot/core"
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
