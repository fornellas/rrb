package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/fornellas/rrb/watcher"
)

var RootCmd = &cobra.Command{
	Use:   "rrb",
	Short: "A filesystem watcher that runs commands on changes",
	Long: "This command watches the filesystem and waits for file changes. When it " +
		"happens, it invokes the given command. If a command was already being previously " +
		"executed, then its process tree is gracefully terminated, before running the " +
		"command again.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			if err := cmd.Help(); err != nil {
				log.Fatal("%w", err)
			}
			os.Exit(1)
		}

		_, err := watcher.NewWatcher(watcher.WatcherConfig{
			RootPath: directory,
			Pattern:  pattern,
		})
		if err != nil {
			log.Fatalf("NewWatcher: %s", err.Error())
		}
		<-make(chan struct{})

	},
}

var directory string
var pattern string

// var ignorePattern string
// var wait float32

// func cobraInit(
// // TODO setup log
// )

func init() {
	// cobra.OnInitialize(cobraInit)
	RootCmd.Flags().StringVarP(
		&directory, "directory", "d", ".",
		"Root directory where to watch for file changes",
	)
	RootCmd.Flags().StringVarP(
		&pattern, "pattern", "p", "**/*.{c,h,cpp,go,py,rb}",
		"Pattern of files to watch for",
	)
	// RootCmd.Flags().StringVarP(
	// 	&ignorePattern, "ignore-pattern", "i", "coverage/*",
	// 	"Pattern of files to ignore",
	// )
	// RootCmd.Flags().Float32VarP(
	// 	&wait, "wait", "w", 3.0,
	// 	"Seconds to wait after SIGTERM before sending SIGKILL",
	// )
}
