package fileprocessing

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"file-counter/internal/types"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"
	"github.com/pterm/pterm"
)

var semaphore = make(chan struct{}, runtime.NumCPU())

func GetSubdirectoriesFileCount(options *types.CommandOptions, fileType commonTypes.FileType) ([]types.DirectoryResult, int64, int, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Counting files...")

	options.RootDirectory = filepath.Clean(options.RootDirectory)

	totalFiles := 0
	var countFiles func(string, bool) error
	countFiles = func(dir string, isVideoRoot bool) error {
		// Acquire a slot in the semaphore
		semaphore <- struct{}{}
		defer func() {
			// Release the slot when the function returns
			<-semaphore
		}()

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

	progressBar, _ := pterm.DefaultProgressbar.WithTotal(totalFiles).WithTitle("Processing files").WithRemoveWhenDone(true).Start()

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
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	var processDir func(string, bool)
	processDir = func(dir string, isVideoRoot bool) {
		defer wg.Done()

		entries, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsPermission(err) {
				errChan <- fmt.Errorf("error reading directory %s: %w", dir, err)
			}
			return
		}

		dirCount := 0
		dirSize := int64(0)

		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				if options.OnlyVideoRoot {
					if entry.Name() == "Videos" {
						wg.Add(1)
						go processDir(path, true)
					} else if !isVideoRoot {
						wg.Add(1)
						go processDir(path, false)
					}
				} else {
					wg.Add(1)
					go processDir(path, isVideoRoot)
				}
				continue
			}

			if options.OnlyVideoRoot && !isVideoRoot {
				progressBar.Increment()
				continue
			}

			info, err := entry.Info()
			if err != nil {
				errChan <- fmt.Errorf("error getting file info for %s: %w", path, err)
				continue
			}

			if commonUtils.IsExtensionValid(fileType, path) {
				dirCount++
				dirSize += info.Size()
			}

			progressBar.Increment()
		}

		if dirCount > 0 {
			mutex.Lock()
			if _, exists := results[dir]; !exists {
				results[dir] = &types.FileInfo{}
			}
			results[dir].Count += dirCount
			results[dir].Size += dirSize
			mutex.Unlock()
		}
	}

	wg.Add(1)
	go processDir(options.RootDirectory, false)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("encountered %d errors during processing: %v", len(errors), errors)
	}

	return results, nil
}
