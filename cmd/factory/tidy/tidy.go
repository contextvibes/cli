// Package tidy provides the command to clean up merged branches.
package tidy

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed tidy.md.tpl
var tidyLongDescription string

// TidyCmd represents the tidy command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var TidyCmd = &cobra.Command{
	Use: "tidy",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("--- Finishing Merged Branch Workflow ---")

		//nolint:exhaustruct // Partial config is sufficient.
		gitClient, err := git.NewClient(
			ctx,
			".",
			git.GitClientConfig{
				Logger:   globals.AppLogger,
				Executor: globals.ExecClient.UnderlyingExecutor(),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to initialize git client: %w", err)
		}

		mainBranch := gitClient.MainBranchName()
		currentBranch, err := gitClient.GetCurrentBranchName(ctx)
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		if currentBranch == mainBranch {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("you are already on the main branch; there is no branch to finish")
		}

		prompt := fmt.Sprintf(
			"This will delete your local branch '%s' and switch to '%s'. Are you sure it has been merged?",
			currentBranch,
			mainBranch,
		)
		confirmed, err := presenter.PromptForConfirmation(prompt)
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			presenter.Info("Aborted by user.")

			return nil
		}

		err = gitClient.SwitchBranch(ctx, mainBranch)
		if err != nil {
			return fmt.Errorf("failed to switch to main branch: %w", err)
		}

		err = gitClient.PullRebase(ctx, mainBranch)
		if err != nil {
			return fmt.Errorf("failed to pull rebase main branch: %w", err)
		}

		err = globals.ExecClient.Execute(ctx, ".", "git", "branch", "-d", currentBranch)
		if err != nil {
			presenter.Warning(
				"Could not delete branch with '-d' (likely not fully merged). Trying '-D'...",
			)
			errForce := globals.ExecClient.Execute(ctx, ".", "git", "branch", "-D", currentBranch)
			if errForce != nil {
				return fmt.Errorf("failed to force delete branch: %w", errForce)
			}
		}
		presenter.Success(
			"Successfully cleaned up '%s' and updated '%s'.",
			currentBranch,
			mainBranch,
		)

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(tidyLongDescription, nil)
	if err != nil {
		panic(err)
	}

	TidyCmd.Short = desc.Short
	TidyCmd.Long = desc.Long
}
