package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/huh"
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

var bar = progress.New(progress.WithDefaultGradient())

func printProgress(current, total float64) {
	if total == 0 {
		return
	}

	pct := current / total
	if pct > 1 {
		pct = 1
	}

	// \r goes back to start of line — rewrites in place instead of new lines
	fmt.Printf("\r  %s", bar.ViewAs(pct))
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
			newFilePath = file + ".mp4"
		}
		newFilePath = strings.ReplaceAll(newFilePath, " ", "_")

		err := convert.ConvertToH264(file, newFilePath, crf, preset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError converting video: %s: %v\n", file, err)
			continue
		}
		tmpFilePath := filepath.Join(tmpFolder, filepath.Base(file))
		err = os.Rename(file, tmpFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError moving file: %s: %v\n", file, err)
			continue
		}
	}
	printProgress(float64(length), float64(length))
	fmt.Println()
	fmt.Println("Conversion complete")
	fmt.Printf("Old files are in: %s — you can delete them if you want\n", tmpFolder)
	return nil
}

var presets = []string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"}

func expandHome(path string) string {
	if path == "~" || path == "~/" {
		home, _ := os.UserHomeDir()
		return home + "/"
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return home + "/" + path[2:]
	}
	return path
}

type config struct {
	folder    string
	recursive bool
	crf       int
	preset    string
}

func promptConfig() (config, error) {
	var folder, crfStr, preset string
	var recursive bool

	crfStr = "18"
	preset = "medium"

	presetOptions := make([]huh.Option[string], len(presets))
	for i, p := range presets {
		presetOptions[i] = huh.NewOption(p, p)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Folder to scan").
				Value(&folder).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("folder is required")
					}
					if _, err := os.Stat(expandHome(s)); err != nil {
						return errors.New("folder does not exist")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Scan subfolders recursively?").
				Value(&recursive),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("CRF quality (0–51)").
				Description("Lower = better quality, larger file. 18 is visually near-lossless.").
				Placeholder("18").
				Value(&crfStr).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("CRF value is required")
					}
					n, err := strconv.Atoi(s)
					if err != nil || n < 0 || n > 51 {
						return errors.New("must be an integer between 0 and 51")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Encoding preset").
				Description("Slower = smaller file, same quality.").
				Options(presetOptions...).
				Value(&preset),
		),
	)

	if err := form.Run(); err != nil {
		return config{}, err
	}

	crf, _ := strconv.Atoi(crfStr)

	return config{
		folder:    expandHome(folder),
		recursive: recursive,
		crf:       crf,
		preset:    preset,
	}, nil
}

func main() {
	cfg, err := promptConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	checkDependencies()

	if err := processVideosFiles(cfg.folder, cfg.recursive, cfg.crf, cfg.preset); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
