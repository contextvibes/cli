// Package product provides commands to manage the product.
package product

import (
	"github.com/contextvibes/cli/cmd/product/build"
	"github.com/contextvibes/cli/cmd/product/clean"
	"github.com/contextvibes/cli/cmd/product/codemod"
	"github.com/contextvibes/cli/cmd/product/format"
	"github.com/contextvibes/cli/cmd/product/quality"
	"github.com/contextvibes/cli/cmd/product/run"
	"github.com/contextvibes/cli/cmd/product/test"
	"github.com/spf13/cobra"
)

// NewProductCmd creates and configures the `product` command.
func NewProductCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "product",
		Short:   "Commands for the product you are building (the 'what').",
		Long:    `The product commands provide tools for building, testing, and ensuring the quality of your codebase.`,
		Example: "contextvibes product --help",
		GroupID: "core",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.AddCommand(build.BuildCmd)
	cmd.AddCommand(test.TestCmd)
	cmd.AddCommand(quality.NewQualityCmd())
	cmd.AddCommand(format.FormatCmd)
	cmd.AddCommand(clean.CleanCmd)
	cmd.AddCommand(run.NewRunCmd())
	cmd.AddCommand(codemod.CodemodCmd)

	return cmd
}
