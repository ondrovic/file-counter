package fileprocessing

import (
	"file-counter/internal/types"
	"file-counter/internal/utils"
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"
)

var semaphore = make(chan struct{}, runtime.NumCPU())

func GetSubdirectoriesFileCount(options *types.CommandOptions, fileType commonTypes.FileType) (directoryResults []types.DirectoryResult, totalSize int64, totalCount int, err error) {
	spinner, err := pterm.DefaultSpinner.Start("Counting files...")
	if err != nil {
		return nil, 0, 0, err
	}

	options.RootDirectory = filepath.Clean(options.RootDirectory)

	var totalFiles int
	totalFiles, err = countFiles(options.RootDirectory, false, options)
	if err != nil {
		spinner.Fail("Error counting files")
		return nil, 0, 0, err
	}

	spinner.Success("File counting complete")

	var progressBar *pterm.ProgressbarPrinter
	progressBar, err = pterm.DefaultProgressbar.WithTotal(totalFiles).WithTitle("Processing files").WithRemoveWhenDone(true).Start()
	if err != nil {
		return nil, 0, 0, err
	}

	var results map[string]*types.FileInfo
	results, err = processDirectory(options, fileType, progressBar)
	if err != nil {
		return nil, 0, 0, err
	}

	directoryResults = make([]types.DirectoryResult, 0, len(results))

	for dir, info := range results {
		directoryResults = append(directoryResults, types.DirectoryResult{
			Directory: dir,
			FileInfo:  *info,
		})
		totalSize += info.Size
		totalCount += info.Count
	}

	sortDirectoryResults(directoryResults, options)

	if _, err = progressBar.Stop(); err != nil {
		return nil, 0, 0, err
	}

	return directoryResults, totalSize, totalCount, nil
}

func processDirectory(options *types.CommandOptions, fileType commonTypes.FileType, progressBar *pterm.ProgressbarPrinter) (map[string]*types.FileInfo, error) {
	results := make(map[string]*types.FileInfo)
	var mutex sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	wg.Add(1)
	go processDir(options.RootDirectory, false, options, fileType, &wg, progressBar, results, &mutex, errChan)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return results, collectErrors(errChan)
}

func processDirEntries(entries []os.DirEntry, dir string, isVideoRoot bool, options *types.CommandOptions, fileType commonTypes.FileType, wg *sync.WaitGroup, progressBar *pterm.ProgressbarPrinter, results map[string]*types.FileInfo, mutex *sync.Mutex, errChan chan<- error) (types.FileInfo, error) {
	var dirInfo types.FileInfo

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if shouldProcessSubDir(entry.Name(), isVideoRoot, options) {
				wg.Add(1)
				go processDir(path, (isVideoRoot || options.OnlyVideoRoot && entry.Name() == options.FilterName), options, fileType, wg, progressBar, results, mutex, errChan)
			}
			continue
		}

		if options.OnlyVideoRoot && !isVideoRoot {
			progressBar.Increment()
			continue
		}

		fileInfo, err := processFile(entry, path, fileType, options)
		if err != nil {
			return dirInfo, err
		}

		dirInfo.Count += fileInfo.Count
		dirInfo.Size += fileInfo.Size
		progressBar.Increment()
	}

	return dirInfo, nil
}

func processDir(dir string, isVideoRoot bool, options *types.CommandOptions, fileType commonTypes.FileType, wg *sync.WaitGroup, progressBar *pterm.ProgressbarPrinter, results map[string]*types.FileInfo, mutex *sync.Mutex, errChan chan<- error) {
	defer wg.Done()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsPermission(err) {
			errChan <- fmt.Errorf("error reading directory %s: %w", dir, err)
		}
		return
	}

	dirInfo, err := processDirEntries(entries, dir, isVideoRoot, options, fileType, wg, progressBar, results, mutex, errChan)
	if err != nil {
		errChan <- err
		return
	}

	if dirInfo.Count > 0 {
		mutex.Lock()
		if _, exists := results[dir]; !exists {
			results[dir] = &types.FileInfo{}
		}
		results[dir].Count += dirInfo.Count
		results[dir].Size += dirInfo.Size
		mutex.Unlock()
	}
}

func shouldProcessSubDir(name string, isVideoRoot bool, options *types.CommandOptions) bool {
	if options.OnlyVideoRoot {
		return name == options.FilterName || !isVideoRoot
	}
	return true
}

func processFile(entry os.DirEntry, path string, fileType commonTypes.FileType, options *types.CommandOptions) (types.FileInfo, error) {
	info, err := entry.Info()
	if err != nil {
		return types.FileInfo{}, fmt.Errorf("error getting file info for %s: %w", path, err)
	}

	// Apply filename filter if provided

	if options.FilterName != "" && (!utils.Contains(utils.ToLower(path), utils.ToLower(options.FilterName))) {
		return types.FileInfo{}, nil // Skip the file if it doesn't match the filter
	}

	if commonUtils.IsExtensionValid(fileType, path) {
		return types.FileInfo{Count: 1, Size: info.Size()}, nil
	}

	return types.FileInfo{}, nil
}

func countFiles(dir string, isVideoRoot bool, options *types.CommandOptions) (int, error) {
	semaphore <- struct{}{}
	defer func() {
		<-semaphore
	}()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsPermission(err) {
			return 0, nil // Skip this directory if permission is denied
		}
		return 0, err
	}

	totalFiles := 0
	for _, entry := range entries {
		if entry.IsDir() {
			newIsVideoRoot := isVideoRoot
			if options.OnlyVideoRoot {
				newIsVideoRoot = true
			}
			count, err := countFiles(filepath.Join(dir, entry.Name()), newIsVideoRoot, options)
			if err != nil {
				return 0, err
			}
			totalFiles += count
		} else if !options.OnlyVideoRoot || isVideoRoot {
			totalFiles++
		}
	}

	return totalFiles, nil
}

func collectErrors(errChan <-chan error) error {
	errors := make([]error, 0, len(errChan)) // Pre-allocate based on the channel's capacity

	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during processing: %v", len(errors), errors)
	}

	return nil
}

func sortDirectoryResults(directoryResults []types.DirectoryResult, options *types.CommandOptions) {
	sort.Slice(directoryResults, func(i, j int) bool {
		var less bool
		switch utils.ToLower(options.SortColumn) {
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
}
