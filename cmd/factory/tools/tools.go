// Package tools provides the command to manage the development toolchain.
package tools

import (
	_ "embed"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

//go:embed tools.md.tpl
var toolsLongDescription string

// ToolsCmd represents the tools command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Force rebuilds and installs development tools (fixes Nix version mismatch).",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		runner := workflow.NewRunner(presenter, globals.AssumeYes)

		return runner.Run(
			ctx,
			"Updating Development Toolchain",
			&workflow.CheckGoEnvStep{
				ExecClient: globals.ExecClient,
				Presenter:  presenter,
			},
			&workflow.ConfigurePathStep{
				Presenter: presenter,
				AssumeYes: globals.AssumeYes,
			},
			&workflow.InstallGoToolsStep{
				ExecClient: globals.ExecClient,
				Presenter:  presenter,
			},
		)
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(toolsLongDescription, nil)
	if err != nil {
		panic(err)
	}

	ToolsCmd.Short = desc.Short
	ToolsCmd.Long = desc.Long
}
