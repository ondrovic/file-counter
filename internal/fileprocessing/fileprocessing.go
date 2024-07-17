package fileprocessing

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
	"file-counter/internal/types"
)

func isFileType(ext string, fileType types.FileType) bool {
	return types.FileExtensions[fileType][strings.ToLower(ext)]
}

func shouldCountFile(ext string, fileType types.FileType) bool {
	return fileType == types.Any || isFileType(ext, fileType)
}

func processDirectory(rootDir string, onlyRoot bool, fileType types.FileType, bar *progressbar.ProgressBar) (map[string]*types.FileInfo, error) {
	results := make(map[string]*types.FileInfo)
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
				results[countDir] = &types.FileInfo{}
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

func GetSubdirectoriesFileCount(rootDir string, sortDescending bool, sortColumn string, onlyRoot bool, fileType types.FileType) ([]types.DirectoryResult, int64, int, error) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "counting "
	s.Suffix = " "
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

	bar := progressbar.NewOptions64(int64(totalFiles), progressbar.OptionSetDescription("building results "), progressbar.OptionFullWidth(), progressbar.OptionClearOnFinish())

	defer bar.Close()

	results, err := processDirectory(rootDir, onlyRoot, fileType, bar)
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