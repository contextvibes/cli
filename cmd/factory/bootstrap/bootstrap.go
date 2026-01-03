// Package bootstrap provides the command for environment initialization.
package bootstrap

import (
	_ "embed"
	"fmt"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

//go:embed bootstrap.md.tpl
var bootstrapLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var installRef string

// BootstrapCmd represents the factory bootstrap command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var BootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Minimal installer for the ContextVibes CLI.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		// ABSOLUTE MINIMAL STEPS: Just the PATH and the Binary.
		steps := []workflow.Step{
			&workflow.ConfigurePathStep{
				Presenter: presenter,
				AssumeYes: globals.AssumeYes,
			},
			&workflow.InstallSelfStep{
				ExecClient: globals.ExecClient,
				Ref:        installRef,
			},
		}

		runner := workflow.NewRunner(presenter, globals.AssumeYes)
		err := runner.Run(ctx, "ContextVibes Minimal Install", steps...)
		if err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}

		presenter.Newline()
		presenter.Success("ContextVibes is now installed!")
		presenter.Header("--- NEXT STEPS ---")
		presenter.Info("1. Refresh shell:  source ~/.bashrc")
		presenter.Info("2. Setup tools:    contextvibes factory tools")
		presenter.Info("3. Setup project:  contextvibes factory scaffold idx")
		presenter.Newline()

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(bootstrapLongDescription, nil)
	if err != nil {
		panic(err)
	}

	BootstrapCmd.Short = desc.Short
	BootstrapCmd.Long = desc.Long

	BootstrapCmd.Flags().StringVar(&installRef, "ref", "main", "Git reference (branch/tag/hash) to install")
}
