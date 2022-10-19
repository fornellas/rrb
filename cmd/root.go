package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	// Use is the one-line usage message.
	// Recommended syntax is as follow:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Use: "rrb",

	// Aliases is an array of aliases that can be used instead of the first word in Use.
	// Aliases []string

	// SuggestFor is an array of command names for which this command will be suggested -
	// similar to aliases but only suggests.
	// SuggestFor []string

	// Short is the short description shown in the 'help' output.
	Short: "A filesystem watcher that runs commands on changes",

	// Long is the long message shown in the 'help <this-command>' output.
	Long: "This command watches the filesystem and waits for file changes. When it " +
		"happens, it invokes the given command. If a command was already being previously " +
		"executed, then its process tree is gracefully terminated, before running the " +
		"command again.",

	// Example is examples of how to use the command.
	// Example string

	// ValidArgs is list of all valid non-flag arguments that are accepted in shell completions
	// ValidArgs []string
	// ValidArgsFunction is an optional function that provides valid non-flag arguments for shell completion.
	// It is a dynamic version of using ValidArgs.
	// Only one of ValidArgs and ValidArgsFunction can be used for a command.
	// ValidArgsFunction func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective)

	// Expected arguments
	// Args PositionalArgs

	// ArgAliases is List of aliases for ValidArgs.
	// These are not suggested to the user in the shell completion,
	// but accepted if entered manually.
	// ArgAliases []string

	// BashCompletionFunction is custom bash functions used by the legacy bash autocompletion generator.
	// For portability with other shells, it is recommended to instead use ValidArgsFunction
	// BashCompletionFunction string

	// Deprecated defines, if this command is deprecated and should print this string when used.
	// Deprecated string

	// Annotations are key/value pairs that can be used by applications to identify or
	// group commands.
	// Annotations map[string]string

	// Version defines the version for this command. If this value is non-empty and the command does not
	// define a "version" flag, a "version" boolean flag will be added to the command and, if specified,
	// will print content of the "Version" variable. A shorthand "v" flag will also be added if the
	// command does not define one.
	// Version string

	// The *Run functions are executed in the following order:
	//   * PersistentPreRun()
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
	// All functions get the same args, the arguments after the command name.
	//
	// PersistentPreRun: children of this command will inherit and execute.
	// PersistentPreRun func(cmd *Command, args []string)
	// PersistentPreRunE: PersistentPreRun but returns an error.
	// PersistentPreRunE func(cmd *Command, args []string) error
	// PreRun: children of this command will not inherit.
	// PreRun func(cmd *Command, args []string)
	// PreRunE: PreRun but returns an error.
	// PreRunE func(cmd *Command, args []string) error
	// Run: Typically the actual work function. Most commands will only implement this.
	// Run func(cmd *Command, args []string)
	// RunE: Run but returns an error.
	// RunE func(cmd *Command, args []string) error
	// PostRun: run after the Run command.
	// PostRun func(cmd *Command, args []string)
	// PostRunE: PostRun but returns an error.
	// PostRunE func(cmd *Command, args []string) error
	// PersistentPostRun: children of this command will inherit and execute after PostRun.
	// PersistentPostRun func(cmd *Command, args []string)
	// PersistentPostRunE: PersistentPostRun but returns an error.
	// PersistentPostRunE func(cmd *Command, args []string) error

	// FParseErrWhitelist flag parse errors to be ignored
	// FParseErrWhitelist FParseErrWhitelist

	// CompletionOptions is a set of options to control the handling of shell completion
	// CompletionOptions CompletionOptions

	// TraverseChildren parses flags on all parents before executing child command.
	// TraverseChildren bool

	// Hidden defines, if this command is hidden and should NOT show up in the list of available commands.
	// Hidden bool

	// SilenceErrors is an option to quiet errors down stream.
	// SilenceErrors bool

	// SilenceUsage is an option to silence usage when an error occurs.
	// SilenceUsage bool

	// DisableFlagParsing disables the flag parsing.
	// If this is true all flags will be passed to the command as arguments.
	// DisableFlagParsing bool

	// DisableAutoGenTag defines, if gen tag ("Auto generated by spf13/cobra...")
	// will be printed by generating docs for this command.
	// DisableAutoGenTag bool

	// DisableFlagsInUseLine will disable the addition of [flags] to the usage
	// line of a command when printing help or generating docs
	// DisableFlagsInUseLine bool

	// DisableSuggestions disables the suggestions based on Levenshtein distance
	// that go along with 'unknown command' messages.
	// DisableSuggestions bool

	// SuggestionsMinimumDistance defines minimum levenshtein distance to display suggestions.
	// Must be > 0.
	// SuggestionsMinimumDistance int
	// contains filtered or unexported fields
}
