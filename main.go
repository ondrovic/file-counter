package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

// FileType represents different types of files
type FileType int

const (
	Any FileType = iota
	Video
	Photo
	Archive
)

// FileInfo stores information about files in a directory
type FileInfo struct {
	Count int
	Size  int64
}

// DirectoryResult represents the result for a directory
type DirectoryResult struct {
	Directory string
	FileInfo
}

var (
	sizeUnits = []struct {
		Name  string
		Value int64
	}{
		{"PB", 1 << 50},
		{"TB", 1 << 40},
		{"GB", 1 << 30},
		{"MB", 1 << 20},
		{"KB", 1 << 10},
		{"B", 1},
	}

	fileExtensions = map[FileType]map[string]bool{
		Video: {
			".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true,
			".flv": true, ".webm": true, ".m4v": true, ".mpg": true, ".mpeg": true,
			".ts": true,
		},
		Photo: {
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
			".tiff": true, ".webp": true, ".svg": true, ".raw": true, ".heic": true,
			".ico": true,
		},
		Archive: {
			".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
			".bz2": true, ".xz": true, ".iso": true, ".tgz": true, ".tbz2": true,
		},
	}
)

func clearConsole() {
	var clearCmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		clearCmd = exec.Command("clear")
	case "windows":
		clearCmd = exec.Command("cmd", "/c", "cls")
	default:
		fmt.Println("Unsupported platform")
		return
	}

	clearCmd.Stdout = os.Stdout
	clearCmd.Run()
}

func formatFileSize(size int64) string {
	for _, unit := range sizeUnits {
		if size >= unit.Value {
			return fmt.Sprintf("%d %s", size/unit.Value, unit.Name)
		}
	}
	return "0 B"
}

func isFileType(ext string, fileType FileType) bool {
	return fileExtensions[fileType][strings.ToLower(ext)]
}

func shouldCountFile(ext string, fileType FileType) bool {
	return fileType == Any || isFileType(ext, fileType)
}

func processDirectory(rootDir string, onlyRoot bool, fileType FileType, bar *progressbar.ProgressBar) (map[string]*FileInfo, error) {
	results := make(map[string]*FileInfo)
	var mutex sync.Mutex

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		dir := filepath.Dir(path)
		countDir := dir
		if onlyRoot {
			countDir = rootDir
		}

		ext := filepath.Ext(path)
		if shouldCountFile(ext, fileType) {
			mutex.Lock()
			if _, exists := results[countDir]; !exists {
				results[countDir] = &FileInfo{}
			}
			results[countDir].Count++
			results[countDir].Size += info.Size()
			mutex.Unlock()
		}

		bar.Add(1)
		return nil
	})

	return results, err
}

func getSubdirectoriesFileCount(rootDir string, sortDescending bool, sortColumn string, onlyRoot bool, fileType FileType) ([]DirectoryResult, int64, int, error) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "counting "
	s.Suffix = "  "
	s.Start()

	rootDir = filepath.Clean(rootDir)

	// Count total files for progress bar
	totalFiles := 0
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFiles++
		}
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}
	s.Stop()

	bar := progressbar.NewOptions64(int64(totalFiles), progressbar.OptionSetDescription("processing"), progressbar.OptionFullWidth(), progressbar.OptionClearOnFinish())

	defer bar.Close()

	results, err := processDirectory(rootDir, onlyRoot, fileType, bar)
	if err != nil {
		return nil, 0, 0, err
	}

	var totalSize int64
	var totalCount int
	directoryResults := make([]DirectoryResult, 0, len(results))

	for dir, info := range results {
		directoryResults = append(directoryResults, DirectoryResult{
			Directory: dir,
			FileInfo:  *info,
		})
		totalSize += info.Size
		totalCount += info.Count
	}

	sort.Slice(directoryResults, func(i, j int) bool {
		var less bool
		switch sortColumn {
		case "Directory":
			less = directoryResults[i].Directory < directoryResults[j].Directory
		case "Size":
			less = directoryResults[i].Size < directoryResults[j].Size
		case "Count":
			less = directoryResults[i].Count < directoryResults[j].Count
		default:
			return false
		}

		if sortDescending {
			return !less
		}
		return less
	})

	return directoryResults, totalSize, totalCount, nil
}

func main() {
	clearConsole()

	app := &cli.App{
		Name:  "subdir-info",
		Usage: "Get file count and total size for each subdirectory in a directory.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "root_dir",
				Aliases:  []string{"r"},
				Usage:    "The root directory to process.",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "sort_descending",
				Aliases: []string{"d"},
				Usage:   "Whether to sort results by count in descending order.",
			},
			&cli.StringFlag{
				Name:    "sort_column",
				Aliases: []string{"s"},
				Value:   "Count",
				Usage:   "The column to sort results by.",
			},
			&cli.BoolFlag{
				Name:     "only_root",
				Aliases:  []string{"o"},
				Required: false,
				Usage:    "Will only count files in the root directory",
			},
			&cli.StringFlag{
				Name:    "file_type",
				Aliases: []string{"t"},
				Value:   "any",
				Usage:   "File type to count (any, video, photo, archive)",
			},
		},
		Action: func(c *cli.Context) error {
			rootDir := c.String("root_dir")
			sortDescending := c.Bool("sort_descending")
			sortColumn := c.String("sort_column")
			onlyRoot := c.Bool("only_root")
			fileTypeStr := c.String("file_type")

			fileType := Any
			switch fileTypeStr {
			case "video":
				fileType = Video
			case "photo":
				fileType = Photo
			case "archive":
				fileType = Archive
			}

			results, totalSize, totalCount, err := getSubdirectoriesFileCount(rootDir, sortDescending, sortColumn, onlyRoot, fileType)
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}

			if len(results) == 0 {
				fmt.Println("No Results Found")
				return nil
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Directory", "Count", "Size"})
			for _, result := range results {
				t.AppendRow(table.Row{
					result.Directory,
					result.Count,
					formatFileSize(result.Size),
				})
			}
			t.AppendFooter(table.Row{"Total", totalCount, formatFileSize(totalSize)})
			t.SetStyle(table.StyleColoredDark)
			t.Render()

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
