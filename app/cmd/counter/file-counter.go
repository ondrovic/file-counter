package main

import (
	"file-counter/internal/fileprocessing"
	"file-counter/internal/types"
	"file-counter/internal/utils"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"
)

var (
	options     = types.CommandOptions{}
	application commonTypes.Application
	version     string

	rootCmd *cobra.Command
)

func registerBoolFlag(cmd *cobra.Command, name, shorthand string, value bool, usage string, target *bool) {
	if !value {
		usage += "\n (default false)"
	} else {
		usage += "\n"
	}
	cmd.Flags().BoolVarP(target, name, shorthand, value, usage)
}

func registerStringFlag(cmd *cobra.Command, name, shorthand, value, usage string, target *string, completionFunc func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) {
	cmd.Flags().StringVarP(target, name, shorthand, value, usage+"\n")

	if completionFunc != nil {
		err := cmd.RegisterFlagCompletionFunc(name, completionFunc)
		if err != nil {
			pterm.Error.Print(err)
		}
	}
}

// func newCompletionCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "completion [bash|zsh|fish|powershell]",
// 		Short: "Generate completion script",
// 		Long:  "To load completions: ...", // Simplified for brevity
// 		Run: func(cmd *cobra.Command, args []string) {
// 			switch args[0] {
// 			case "bash":
// 				cmd.Root().GenBashCompletion(os.Stdout)
// 			case "zsh":
// 				cmd.Root().GenZshCompletion(os.Stdout)
// 			case "fish":
// 				cmd.Root().GenFishCompletion(os.Stdout, true)
// 			case "powershell":
// 				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
// 			}
// 		},
// 	}
// }

func init() {

	appName := "File-Counter"

	application = commonTypes.Application{
		Name:        appName,
		Description: "File counting cli tool",
		Style: commonTypes.Styles{
			Color: commonTypes.Colors{
				Background: pterm.BgDarkGray,
				Foreground: pterm.FgWhite,
			},
		},
		Usage:   pterm.Sprintf("%s [root-directory]", utils.ToLower(appName)),
		Version: commonUtils.GetVersion(version, "0.0.1-dev"),
	}

	rootCmd = &cobra.Command{
		Use:     application.Usage,
		Short:   application.Description,
		Args:    cobra.MinimumNArgs(1),
		Run:     runCounter,
		Version: application.Version,
	}

	rootCmd.SetVersionTemplate(`{{printf "Version: %s\n" .Version}}`)

	registerBoolFlag(rootCmd, "sort-descending", "d", false, "Whether to sort results by count in descending order", &options.SortDescending)
	registerStringFlag(rootCmd, "sort-column", "s", "Count", "The column to sort results by\n (Choices: count, directory, size)", &options.SortColumn, nil)
	registerBoolFlag(rootCmd, "only-video-root", "o", false, "Only count files in the root of Videos folders", &options.OnlyVideoRoot)
	registerStringFlag(rootCmd, "file-type", "t", string(commonTypes.FileTypes.Any), "File type to count\n (Choices: any, video, image, archive, documents)", &options.FileType, nil)
	registerStringFlag(rootCmd, "filter-name", "n", "", "Name to filter by, filters both file and directory", &options.FilterName, nil)

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		pterm.Error.Print(err)
	}

}

func runCounter(_ *cobra.Command, args []string) {

	options.RootDirectory = args[0]

	fileType := commonUtils.ToFileType(options.FileType)

	if options.OnlyVideoRoot {
		options.FilterName = "Videos"
	}

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
	if err := commonUtils.ApplicationBanner(&application, commonUtils.ClearTerminalScreen); err != nil {
		pterm.Error.Print(err)
		return
	}

	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		return
	}
}
