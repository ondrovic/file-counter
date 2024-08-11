package main

import (
	"os"
	"runtime"

	"file-counter/internal/fileprocessing"
	"file-counter/internal/types"
	"file-counter/internal/utils"

	commonUtils "github.com/ondrovic/common/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "file-counter [root-directory]",
	Short: "A file counting application",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

var (
	options = types.CommandOptions{}
)

func registerBoolFlag(cmd *cobra.Command, name, shorthand string, value bool, usage string, target *bool) {
	if !value {
		usage = usage + "\n (default false)"
	} else {
		usage = usage + "\n"
	}
	cmd.Flags().BoolVarP(target, name, shorthand, value, usage)
}

func registerStringFlag(cmd *cobra.Command, name, shorthand, value, usage string, target *string, completionFunc func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) {
	cmd.Flags().StringVarP(target, name, shorthand, value, usage + "\n")
	if completionFunc != nil {
		cmd.RegisterFlagCompletionFunc(name, completionFunc)
	}
}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long:  "To load completions: ...", // Simplified for brevity
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
}

func init() {
	registerBoolFlag(rootCmd, "sort_descending", "d", false, "Whether to sort results by count in descending order", &options.SortDescending)
	registerStringFlag(rootCmd, "sort_column", "s", "Count", "The column to sort results by", &options.SortColumn, nil)
	registerBoolFlag(rootCmd, "only_video_root", "v", false, "Only count files in the root of Videos folders", &options.OnlyVideoRoot)
	registerStringFlag(rootCmd, "file_type", "t", "any", "File type to count", &options.FilterType, nil)
	
	rootCmd.AddCommand(newCompletionCmd())
}

func run(cmd *cobra.Command, args []string) {
	options.RootDirectory = args[0]

	fileType := commonUtils.ToFileType(options.FilterType)

	results, totalSize, totalCount, err := fileprocessing.GetSubdirectoriesFileCount(&options, fileType)
	if err != nil {
		pterm.Error.Printf("Error: %v\n", err)
	}

	if len(results) == 0 {
		pterm.Info.Printf("%v results found\n", len(results))
		return
	}

	utils.RenderResultsTable(results, totalSize, totalCount)
}

func main() {
	commonUtils.ClearTerminalScreen(runtime.GOOS)
	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Printf("Error: %v\n", err)
	}
}
