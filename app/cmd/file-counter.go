package main

import (
	"fmt"
	"os"
	"strings"

	"file-counter/internal/fileprocessing"
	"file-counter/internal/types"
	"file-counter/internal/utils"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "file-counter",
	Short: "A file counting application",
	Run:   run,
}

var (
	options = types.CommandOptions{}
)

func init() {
	rootCmd.Flags().StringVarP(&options.RootDirectory, "root_dir", "r", "", "The root directory to process")
	rootCmd.MarkFlagRequired("root_dir")

	rootCmd.Flags().BoolVarP(&options.SortDescending, "sort_descending", "d", false, "Whether to sort results by count in descending order")

	rootCmd.Flags().StringVarP(&options.SortColumn, "sort_column", "s", "Count", "The column to sort results by")
	rootCmd.RegisterFlagCompletionFunc("sort_column", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"Directory", "Count", "Size"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.Flags().BoolVarP(&options.OnlyRoot, "only_root", "o", false, "Will only count files in the root directory")

	rootCmd.Flags().StringVarP(&options.FilterType, "file_type", "t", "any", "File type to count")
	rootCmd.RegisterFlagCompletionFunc("file_type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"any", "video", "image", "archive", "documents"}, cobra.ShellCompDirectiveNoFileComp
	})

	rootCmd.Flags().BoolVarP(&options.GroupByParent, "group_parent", "g", true, "Will group folders by parent")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(file-counter completion bash)

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  # Note: your $fpath[1] may not be the correct index based on the plugins you have loaded
        # echo "$fpath" - to located the completions index
		# for oh-my-zsh - echo $fpath | tr ' ' '\n' | grep -n '\.oh-my-zsh/completions$' | awk -F: '{print $1}'
  $ file-counter completion zsh > "${fpath[index_num]}/_file-counter"

Fish:
  $ file-counter completion fish | source

PowerShell:
  PS> file-counter completion powershell | Out-String | Invoke-Expression
`,
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
	})
}

func run(cmd *cobra.Command, args []string) {
	utils.ClearConsole()

	fileType := types.Any
	switch strings.ToLower(options.FilterType) {
	case "video":
	case "videos":
		fileType = types.Video
	case "image":
		fileType = types.Image
	case "archive":
		fileType = types.Archive
	case "documents":
		fileType = types.Documents
	}

	results, totalSize, totalCount, err := fileprocessing.GetSubdirectoriesFileCount(&options, fileType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("No Results Found")
		return
	}

	utils.RenderResultsTable(results, totalSize, totalCount)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
