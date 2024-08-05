package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"file-counter/internal/types"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pterm/pterm"
)

func ClearConsole() {
	var clearCmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		clearCmd = exec.Command("clear")
	case "windows":
		clearCmd = exec.Command("cmd", "/c", "cls")
	default:
		// fmt.Println("Unsupported platform")
		pterm.Info.Println("Unsupported platform")
		return
	}

	clearCmd.Stdout = os.Stdout
	clearCmd.Run()
}

func FormatFileSize(size int64) string {
	for _, unit := range types.SizeUnits {
		if size >= unit.Value {
			return pterm.Sprintf("%d %s", size/unit.Value, unit.Name)
		}
	}
	return "0 B"
}

func formatPath(path string) string {
	switch runtime.GOOS {
    case "windows":
        // Convert to Windows style paths (with backslashes)
        return filepath.ToSlash(path)
    case "linux", "darwin":
        // Convert to Unix style paths (with forward slashes)
        return filepath.FromSlash(path)
    default:
        // Default to Unix style paths
        return path
    } 
}

func RenderResultsTable(results []types.DirectoryResult, totalSize int64, totalCount int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Directory", "Count", "Size"})
	for _, result := range results {
		t.AppendRow(table.Row{
			formatPath(result.Directory),
			result.Count,
			FormatFileSize(result.Size),
		})
	}
	t.AppendFooter(table.Row{"Total", totalCount, FormatFileSize(totalSize)})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}