package files

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

var legacyVideoContainers = []string{
	"avi",
	"asf",
	"mts",
	"m2ts",
	"ts",
	"vob",
	"flv",
	"3gp",
	"3g2",
	"mpg",
	"mpeg",
}

var legacyVideoCodecs = []string{
	"mpeg1video",
	"mpeg2video",
	"mjpeg",
	"theora",
	"svq1",
	"svq3",
	"cinepak",
	"flv1",
	"vp6",
	"dv",
	"h263",
}

type Stream struct {
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`
	PixFmt    string `json:"pix_fmt"`
}

type Format struct {
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
}

type FileFormat struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

func extractDataFromFileFormat(fileName string) FileFormat {
	var fileFormat FileFormat

	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		fileName,
	)

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("ffprobe failed on ", fileName)
		return FileFormat{}
	}

	err = json.Unmarshal(out, &fileFormat)
	if err != nil {
		fmt.Printf("failed to unmarshal file format: %v\n", err)
		return FileFormat{}
	}

	return fileFormat
}

func getContainer(fileFormat FileFormat) string {
	return fileFormat.Format.FormatName
}

func isLegacyContainer(container string) bool {
	return slices.Contains(legacyVideoContainers, container)
}

func getCodec(fileFormat FileFormat) (string, bool) {
	for _, stream := range fileFormat.Streams {
		if stream.CodecType == "video" {
			return stream.CodecName, true
		}
	}
	return "", false
}

func isLegacyCodec(codec string) bool {
	return slices.Contains(legacyVideoCodecs, codec)
}

func ScanFolder(folder string, recursive bool) ([]string, error) {
	fmt.Println("Scanning folder: ", folder)

	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	var legacyVideoPaths []string
	for _, file := range files {
		if file.IsDir() {
			if !recursive {
				continue
			}

			subPaths, err := ScanFolder(filepath.Join(folder, file.Name()), recursive)

			if err != nil {
				return nil, err
			}

			legacyVideoPaths = append(legacyVideoPaths, subPaths...)
			continue
		}
		path := filepath.Join(folder, file.Name())
		videoFormat := extractDataFromFileFormat(path)

		container := getContainer(videoFormat)
		codec, ok := getCodec(videoFormat)
		if !ok {
			fmt.Println("No video codec found for ", path)
			continue
		}

		if isLegacyCodec(codec) || isLegacyContainer(container) {
			legacyVideoPaths = append(legacyVideoPaths, path)
		}
	}

	return legacyVideoPaths, nil
}
