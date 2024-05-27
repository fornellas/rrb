package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/sirupsen/logrus"

	"github.com/fornellas/rrb/log"
	"github.com/fornellas/rrb/runner"
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
		if len(args) == 0 {
			logrus.Fatal("Missing build command.")
			os.Exit(1)
		}

		w, err := watcher.NewWatcher(watcher.Config{
			RootPath:         directory,
			Patterns:         patterns,
			IgnorePatterns:   ignorePatterns,
			DebounceDuration: debounce,
		})
		if err != nil {
			logrus.Fatalf("NewWatcher: %s", err.Error())
		}
		defer w.Close()

		r := runner.NewRunner(killWait, args[0], args[1:]...)

		sigCn := make(chan os.Signal, 1)
		signal.Notify(sigCn, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		for {
			select {
			case <-w.ChangedFilesCn:
				if err := r.Run(); err != nil {
					logrus.Fatal(err)
				}
			case err := <-w.ErrorsCn:
				logrus.Fatal(err)
			case sig := <-sigCn:
				logrus.Warnf("Signal %s received!", sig)
				r.Kill()
				logrus.Fatal("Exiting")
			}
		}
	},
}

var directory string
var patterns []string
var ignorePatterns []string
var debounce time.Duration
var killWait time.Duration
var logLevel string

func cobraInit() {
	if err := log.Setup(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func init() {
	cobra.OnInitialize(cobraInit)
	RootCmd.Flags().StringVarP(
		&directory, "directory", "d", ".",
		"Root directory where to watch for file changes",
	)
	RootCmd.Flags().StringArrayVarP(
		&patterns, "pattern", "p", []string{
			"**/*.{c,h,cpp,go,py,rb,sh,kt}",
		},
		"Pattern to watch for changes (relative to given directory)",
	)
	RootCmd.Flags().StringArrayVarP(
		&ignorePatterns, "ignore-pattern", "i", []string{
			"**/*.{o,a,la,pyc}",
		},
		"Pattern to ignore changes (relative to given directory)",
	)
	RootCmd.Flags().DurationVarP(
		&debounce, "debounce", "b", 500*time.Millisecond,
		"Idle time after file change before calling build",
	)
	RootCmd.Flags().DurationVarP(
		&killWait, "kill-wait", "w", time.Second,
		"Seconds to wait after SIGTERM before sending SIGKILL",
	)
	RootCmd.Flags().StringVarP(
		&logLevel, "log-level", "l", "info",
		"Logging level",
	)
}
