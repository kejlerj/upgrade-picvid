package convert

import (
	"fmt"
	"os/exec"
	"strconv"
)

func ConvertToH264(filePath string, outputPath string, crf int, preset string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", filePath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-crf", strconv.Itoa(crf),
		"-preset", preset,
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		outputPath,
	)

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ffmpeg failed to convert to h264 on %s: %s", filePath, err)
	}

	return nil
}
