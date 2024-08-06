package utils

import (
	"os"
	"runtime"

	"file-counter/internal/types"

	"github.com/jedib0t/go-pretty/v6/table"
	commonUtils "github.com/ondrovic/common/utils"
)

// RenderResultsToTable
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
