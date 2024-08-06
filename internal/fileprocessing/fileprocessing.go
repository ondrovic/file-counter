package fileprocessing

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"file-counter/internal/types"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"
	"github.com/pterm/pterm"
)

func GetSubdirectoriesFileCount(options *types.CommandOptions, fileType commonTypes.FileType) ([]types.DirectoryResult, int64, int, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Counting files...")

	options.RootDirectory = filepath.Clean(options.RootDirectory)

	totalFiles := 0
	var countFiles func(string, bool) error
	countFiles = func(dir string, isVideoRoot bool) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsPermission(err) {
				return nil // Skip this directory
			}
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				newIsVideoRoot := isVideoRoot
				if options.OnlyVideoRoot && entry.Name() == "Videos" {
					newIsVideoRoot = true
				}
				if err := countFiles(filepath.Join(dir, entry.Name()), newIsVideoRoot); err != nil {
					return err
				}
			} else {
				if !options.OnlyVideoRoot || isVideoRoot {
					totalFiles++
				}
			}
		}
		return nil
	}

	err := countFiles(options.RootDirectory, false)
	if err != nil {
		spinner.Fail("Error counting files")
		return nil, 0, 0, err
	}
	spinner.Success("File counting complete")

	progressBar, _ := pterm.DefaultProgressbar.WithTotal(totalFiles).WithTitle("Processing files").Start()

	results, err := processDirectory(options, fileType, progressBar)
	if err != nil {
		return nil, 0, 0, err
	}

	var totalSize int64
	var totalCount int
	directoryResults := make([]types.DirectoryResult, 0, len(results))

	for dir, info := range results {
		directoryResults = append(directoryResults, types.DirectoryResult{
			Directory: dir,
			FileInfo:  *info,
		})
		totalSize += info.Size
		totalCount += info.Count
	}

	sort.Slice(directoryResults, func(i, j int) bool {
		var less bool
		switch strings.ToLower(options.SortColumn) {
		case "directory", "directories":
			less = directoryResults[i].Directory < directoryResults[j].Directory
		case "size":
			less = directoryResults[i].Size < directoryResults[j].Size
		case "count":
			less = directoryResults[i].Count < directoryResults[j].Count
		default:
			return false
		}

		if options.SortDescending {
			return !less
		}
		return less
	})

	progressBar.Stop()

	return directoryResults, totalSize, totalCount, nil
}

func processDirectory(options *types.CommandOptions, fileType commonTypes.FileType, progressBar *pterm.ProgressbarPrinter) (map[string]*types.FileInfo, error) {
	results := make(map[string]*types.FileInfo)
	var mutex sync.Mutex

	var processDir func(string, bool) error
	processDir = func(dir string, isVideoRoot bool) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsPermission(err) {
				return nil // Skip this directory
			}
			return err
		}

		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				if options.OnlyVideoRoot {
					if entry.Name() == "Videos" {
						if err := processDir(path, true); err != nil {
							return err
						}
					} else if !isVideoRoot {
						if err := processDir(path, false); err != nil {
							return err
						}
					}
				} else {
					if err := processDir(path, isVideoRoot); err != nil {
						return err
					}
				}
				continue
			}

			if options.OnlyVideoRoot && !isVideoRoot {
				progressBar.Increment()
				continue
			}

			info, err := entry.Info()
			if err != nil {
				return err
			}

			countDir := dir
			if options.OnlyRoot {
				countDir = options.RootDirectory
			} else if options.GroupByParent {
				countDir = filepath.Dir(countDir)
			}

			if commonUtils.IsExtensionValid(fileType, path) {
				mutex.Lock()
				if _, exists := results[countDir]; !exists {
					results[countDir] = &types.FileInfo{}
				}
				results[countDir].Count++
				results[countDir].Size += info.Size()
				mutex.Unlock()
			}

			progressBar.Increment()
		}
		return nil
	}

	err := processDir(options.RootDirectory, false)
	return results, err
}
