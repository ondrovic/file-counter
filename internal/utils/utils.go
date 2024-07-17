package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/jedib0t/go-pretty/v6/table"
	"file-counter/internal/types"
)

func ClearConsole() {
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

func FormatFileSize(size int64) string {
	for _, unit := range types.SizeUnits {
		if size >= unit.Value {
			return fmt.Sprintf("%d %s", size/unit.Value, unit.Name)
		}
	}
	return "0 B"
}

func RenderResultsTable(results []types.DirectoryResult, totalSize int64, totalCount int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Directory", "Count", "Size"})
	for _, result := range results {
		t.AppendRow(table.Row{
			result.Directory,
			result.Count,
			FormatFileSize(result.Size),
		})
	}
	t.AppendFooter(table.Row{"Total", totalCount, FormatFileSize(totalSize)})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}