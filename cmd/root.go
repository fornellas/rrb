package cmd

import (
	"log"
	"os"
	"time"

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

		w, err := watcher.NewWatcher(watcher.Config{
			RootPath: directory,
			Patterns: patterns,
			// IgnorePatterns:   ignorePatterns,
			DebounceDuration: debounce,
		})
		if err != nil {
			log.Fatalf("NewWatcher: %s", err.Error())
		}
		defer w.Close()

		for {
			select {
			case <-w.ChangedFilesCn:
				log.Println("GOTTA BUILD!!!!")
			case err := <-w.ErrorsCn:
				log.Println("ERROR: ", err)
			}
		}
	},
}

var directory string
var patterns []string
var debounce time.Duration

// var ignorePatterns []string

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
	RootCmd.Flags().StringArrayVarP(
		&patterns, "pattern", "p", []string{
			"**/*.{c,h,cpp,go,py,rb,sh,js,java,kt,php}",
		},
		"Pattern to watch for changes (relative to given directory)",
	)
	RootCmd.Flags().DurationVarP(
		&debounce, "debounce", "b", 200*time.Millisecond,
		"Idle time after file change before calling build",
	)
	// RootCmd.Flags().StringArrayVarP(
	// 	&ignorePatterns, "ignore", "i", []string{
	// 		"/.git/*",
	// 		"/**/*.{o}",
	// 	},
	// 	"usage string",
	// )

	// RootCmd.Flags().Float32VarP(
	// 	&wait, "wait", "w", 3.0,
	// 	"Seconds to wait after SIGTERM before sending SIGKILL",
	// )
}
