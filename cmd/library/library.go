// Package library provides commands to manage the knowledge library.
package library

import (
	"github.com/contextvibes/cli/cmd/library/index"
	"github.com/contextvibes/cli/cmd/library/systemprompt"
	"github.com/contextvibes/cli/cmd/library/thea"
	"github.com/contextvibes/cli/cmd/library/vendor"
	"github.com/spf13/cobra"
)

// NewLibraryCmd creates and returns the base command for the 'library' subcommand group.
func NewLibraryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "library",
		Short: "Commands for knowledge and standards (the 'where').",
		Long: `The library commands provide tools for managing and interacting with the knowledge base,
including THEA standards, system prompts, and indexed documentation.`,
		Example: "contextvibes library --help",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},

		// Boilerplate
		Aliases:                    []string{},
		SuggestFor:                 []string{},
		ValidArgs:                  []string{},
		ValidArgsFunction:          nil,
		Args:                       nil,
		ArgAliases:                 []string{},
		BashCompletionFunction:     "",
		Deprecated:                 "",
		Annotations:                nil,
		Version:                    "",
		PersistentPreRun:           nil,
		PersistentPreRunE:          nil,
		PreRun:                     nil,
		PostRun:                    nil,
		PostRunE:                   nil,
		PersistentPostRun:          nil,
		PersistentPostRunE:         nil,
		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
		CompletionOptions:          cobra.CompletionOptions{},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              true,
		SilenceUsage:               true,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          true,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}

	// Add subcommands
	cmd.AddCommand(index.NewIndexCmd())
	cmd.AddCommand(thea.TheaCmd)
	cmd.AddCommand(systemprompt.SystemPromptCmd)
	cmd.AddCommand(vendor.VendorCmd)

	return cmd
}
