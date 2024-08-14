package types

// FileInfo stores information about files in a directory.
type FileInfo struct {
	Count int
	Size  int64
}

// DirectoryResult represents the result for a directory.
type DirectoryResult struct {
	Directory string
	FileInfo
}

// CommandOptions represents the command line options.
type CommandOptions struct {
	FileType       string
	FilterName     string
	OnlyRoot       bool
	OnlyVideoRoot  bool
	RootDirectory  string
	SortColumn     string
	SortDescending bool
}
