// cmd/finish.go
package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var finishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Pushes the current branch and creates a GitHub pull request.",
	Long: `Standardizes the process of finalizing a feature branch.

This command first pushes the current branch to the remote ('origin' by default).
Then, if the GitHub CLI ('gh') is installed, it interactively creates a pull
request, filling in details and opening it in a web browser.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		logger := AppLogger
		ctx := context.Background()

		presenter.Summary("Finishing work on the current branch.")

		gitClient, err := git.NewClient(
			ctx,
			".",
			git.GitClientConfig{Logger: logger, Executor: ExecClient.UnderlyingExecutor()},
		)
		if err != nil {
			presenter.Error("Failed to initialize Git client: %v", err)
			return err
		}

		currentBranch, err := gitClient.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Error("Could not determine current branch: %v", err)
			return err
		}

		if currentBranch == gitClient.MainBranchName() {
			presenter.Error(
				"You cannot create a pull request from the main branch ('%s').",
				gitClient.MainBranchName(),
			)
			return errors.New("cannot finish from main branch")
		}

		presenter.Info("Current branch is '%s'.", currentBranch)
		pushConfirmed, err := presenter.PromptForConfirmation(
			"Push this branch to the remote repository?",
		)
		if err != nil {
			return err
		}
		if !pushConfirmed {
			presenter.Info("Aborted. Your branch was not pushed.")
			return nil
		}

		presenter.Step("Pushing '%s' to remote...", currentBranch)
		if err := gitClient.Push(ctx, currentBranch); err != nil {
			presenter.Error("Failed to push branch: %v", err)
			return err
		}
		presenter.Success("Branch pushed successfully.")

		if !ExecClient.CommandExists("gh") {
			presenter.Warning("GitHub CLI ('gh') not found in your PATH.")
			presenter.Advice(
				"Cannot create the PR automatically. Please install 'gh' or create the PR manually on GitHub.",
			)
			return nil
		}

		presenter.Newline()
		prConfirmed, err := presenter.PromptForConfirmation("Create a GitHub Pull Request now?")
		if err != nil {
			return err
		}
		if !prConfirmed {
			presenter.Info("Aborted. You can create the PR later by running 'gh pr create'.")
			return nil
		}

		presenter.Step("Running 'gh pr create'...")
		if err := ExecClient.Execute(ctx, ".", "gh", "pr", "create", "--fill", "--web"); err != nil {
			presenter.Error("Failed to create pull request: %v", err)
			return err
		}

		presenter.Success("Pull Request created and opened in your browser.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(finishCmd)
}
