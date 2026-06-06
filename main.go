package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kejlerj/upgrade-picvid/convert"
	"github.com/kejlerj/upgrade-picvid/files"
)

func checkDependencies() {
	for _, bin := range []string{"ffmpeg", "ffprobe"} {
		_, err := exec.LookPath(bin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %q not found. Install it with: brew %s\n", bin, bin)

			fmt.Println("Do you want to install it? (y/n)")
			var input string
			fmt.Scanln(&input)

			if input == "y" {
				err := exec.Command("brew", "install", bin).Run()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to install %s: %v\n", bin, err)
					os.Exit(1)
				}
			} else {
				fmt.Println("Exiting...")
				os.Exit(1)
			}
		}
	}
}

func printProgress(current, total float64) {
	if total == 0 {
		return
	}

	pct := current / total
	if pct > 1 {
		pct = 1
	}

	barWidth := 40
	filled := int(pct * float64(barWidth))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// \r goes back to start of line — rewrites in place instead of new lines
	fmt.Printf("\r  [%s] %3.0f%%", bar, pct*100)
}

func processVideosFiles(folder string, recursive bool, crf int, preset string) error {
	videoFiles, err := files.ScanFolder(folder, recursive)
	if err != nil {
		return fmt.Errorf("error scanning folder: %v", err)
	}

	length := len(videoFiles)

	if length == 0 {
		fmt.Println("No legacy videos found")
		return nil
	}

	fmt.Printf("Converting %d videos\n", length)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting working directory: %v", err)
	}
	tmpFolder := filepath.Join(userHomeDir, "legacy-picvid")
	err = os.MkdirAll(tmpFolder, 0755)
	if err != nil {
		return fmt.Errorf("error creating tmp folder: %v", err)
	}

	for i, file := range videoFiles {
		printProgress(float64(i), float64(length))

		newFilePath := ""
		extension := filepath.Ext(file)
		if extension != "" {
			newFilePath = strings.Replace(file, extension, ".mp4", 1)
		} else {
			newFilePath = strings.Join([]string{file, ".mp4"}, "")
		}
		newFilePath = strings.ReplaceAll(newFilePath, " ", "_")

		err := convert.ConvertToH264(file, newFilePath, crf, preset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting video: %s: %v\n", file, err)
			continue
		}
		tmpFilePath := filepath.Join(tmpFolder, filepath.Base(file))
		err = os.Rename(file, tmpFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error moving file: %s: %v\n", file, err)
			continue
		}
	}
	printProgress(float64(length), float64(length))
	fmt.Println()
	fmt.Println("Conversion complete")
	fmt.Printf("Old files are in: %s you can delete them if you want\n", tmpFolder)
	return nil
}

var presets = []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"}

func main() {
	folder := flag.String("folder", "", "folder to scan")
	recursive := flag.Bool("recursive", false, "scan recursively")
	crf := flag.Int("crf", 18, "CRF value")
	preset := flag.String("preset", "medium", "preset value")

	flag.Parse()

	if *folder == "" {
		fmt.Fprintf(os.Stderr, "Folder is required\n")
		os.Exit(1)
	}
	if *crf < 0 || *crf > 51 {
		fmt.Fprintf(os.Stderr, "CRF value must be between 0 and 51\n")
		os.Exit(1)
	}
	if !slices.Contains(presets, *preset) {
		fmt.Fprintf(os.Stderr, "Preset value must be one of: %s\n", strings.Join(presets, ", "))
		os.Exit(1)
	}

	checkDependencies()

	err := processVideosFiles(*folder, *recursive, *crf, *preset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}
