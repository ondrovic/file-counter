package utils

import (
	"os"
	"runtime"
	"strings"

	"file-counter/internal/types"

	"github.com/jedib0t/go-pretty/v6/table"
	commonUtils "github.com/ondrovic/common/utils"
)

// ToLower converts a string ToLower.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Contains returns if a string contains and item.
func Contains(s, i string) bool {
	return strings.Contains(s, i)
}

// RenderResultsToTable.
func RenderResultsTable(results []types.DirectoryResult, totalSize int64, totalCount int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Directory", "Count", "Size"})
	for _, result := range results {
		t.AppendRow(table.Row{
			commonUtils.FormatPath(result.Directory, runtime.GOOS),
			result.Count,
			commonUtils.FormatSize(result.Size),
		})
	}
	t.AppendFooter(table.Row{"Total", totalCount, commonUtils.FormatSize(totalSize)})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}
