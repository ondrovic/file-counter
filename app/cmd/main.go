package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"file-counter/internal/fileprocessing"
	"file-counter/internal/types"
	"file-counter/internal/utils"
)

func main() {
	utils.ClearConsole()

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
				Usage:   "File type to count (any, video, image, archive)",
			},
		},
		Action: func(c *cli.Context) error {
			rootDir := c.String("root_dir")
			sortDescending := c.Bool("sort_descending")
			sortColumn := c.String("sort_column")
			onlyRoot := c.Bool("only_root")
			fileTypeStr := c.String("file_type")

			fileType := types.Any
			switch fileTypeStr {
			case "video":
				fileType = types.Video
			case "image":
				fileType = types.Image
			case "archive":
				fileType = types.Archive
			}

			results, totalSize, totalCount, err := fileprocessing.GetSubdirectoriesFileCount(rootDir, sortDescending, sortColumn, onlyRoot, fileType)
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}

			if len(results) == 0 {
				fmt.Println("No Results Found")
				return nil
			}

			utils.RenderResultsTable(results, totalSize, totalCount)

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}