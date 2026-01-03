// Package squash provides the command to squash commits on a feature branch.
package squash

import (
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

// SquashCmd represents the squash command.
var SquashCmd = &cobra.Command{
	Use:   "squash",
	Short: "Squashes all commits on the current feature branch into one.",
	Long: `Performs a "Soft Reset" to the merge-base of the main branch.
This stages all changes from your multiple commits into a single pending commit.
It automatically generates '_contextvibes.md' containing the diff, allowing
you to use an AI to generate the summary message before committing.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		//nolint:exhaustruct // Partial config is sufficient.
		gitCfg := git.GitClientConfig{
			Logger:   globals.AppLogger,
			Executor: globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, cwd, gitCfg)
		if err != nil {
			return fmt.Errorf("failed to initialize git client: %w", err)
		}

		// Initialize Workflow State
		//nolint:exhaustruct // State is populated by steps.
		state := &workflow.SquashState{}

		runner := workflow.NewRunner(presenter, globals.AssumeYes)

		return runner.Run(
			ctx,
			"Squashing Feature Branch",
			&workflow.EnsureCleanOrSaveStep{
				GitClient: client,
				Presenter: presenter,
				AssumeYes: globals.AssumeYes,
			},
			&workflow.AnalyzeBranchStep{
				GitClient: client,
				Presenter: presenter,
				State:     state,
			},
			&workflow.SoftResetStep{
				GitClient: client,
				Presenter: presenter,
				State:     state,
				AssumeYes: globals.AssumeYes,
			},
			&workflow.GenerateSquashPromptStep{
				GitClient: client,
				Presenter: presenter,
				State:     state,
			},
			&workflow.CommitSquashStep{
				GitClient: client,
				Presenter: presenter,
				State:     state,
				AssumeYes: globals.AssumeYes,
			},
			&workflow.ForcePushStep{
				GitClient: client,
				Presenter: presenter,
				State:     state,
				AssumeYes: globals.AssumeYes,
			},
		)
	},
}
