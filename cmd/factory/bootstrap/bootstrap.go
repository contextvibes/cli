// Package bootstrap provides the command for environment initialization.
package bootstrap

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

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
	Short: "Fast-track installation and environment setup.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		// Detect if we are running via 'go run' (temporary) or as a local binary
		exePath, _ := os.Executable()
		isGoRun := strings.Contains(exePath, "go-build")

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

		// If we are already local, we can do the heavy lifting immediately.
		// If we are 'go run', we stop after installation to let the user refresh their shell.
		if !isGoRun {
			steps = append(steps,
				&workflow.InstallGoToolsStep{ExecClient: globals.ExecClient, Presenter: presenter},
				&workflow.ScaffoldIDXStep{Presenter: presenter, AssumeYes: globals.AssumeYes},
			)
		}

		runner := workflow.NewRunner(presenter, globals.AssumeYes)
		err := runner.Run(ctx, "ContextVibes Bootstrap", steps...)
		if err != nil {
			return fmt.Errorf("bootstrap workflow failed: %w", err)
		}

		if isGoRun {
			presenter.Newline()
			presenter.Success("Step 1 Complete: ContextVibes is installed!")
			presenter.Header("--- FINAL ACTION REQUIRED ---")
			presenter.Info("Run the following to refresh your shell and finish setup:")
			presenter.Detail("source ~/.bashrc && contextvibes factory bootstrap")
			presenter.Newline()
		} else {
			presenter.Success("Environment is fully configured and ready for use.")
		}

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
