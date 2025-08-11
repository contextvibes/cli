// cmd/gittidy.go
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var gitTidyCmd = &cobra.Command{
	Use:   "git-tidy",
	Short: "Provides tools for Git branch hygiene.",
	Long: `Provides interactive subcommands for cleaning up local Git branches,
such as deleting a branch after its pull request has been merged.`,
}

var gitTidyFinishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Deletes the current branch and switches to the main branch.",
	Long: `A workflow for after your pull request has been merged.
It safely switches to the main branch, pulls the latest changes, and then
deletes the local feature branch you were on.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		logger := AppLogger
		ctx := cmd.Context()

		presenter.Summary("--- Finishing Merged Branch Workflow ---")

		gitClient, err := git.NewClient(ctx, ".", git.GitClientConfig{Logger: logger, Executor: ExecClient.UnderlyingExecutor()})
		if err != nil {
			presenter.Error("Failed to initialize Git client: %v", err)
			return err
		}

		mainBranch := gitClient.MainBranchName()
		currentBranch, err := gitClient.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Error("Could not determine current branch: %v", err)
			return err
		}
		if currentBranch == mainBranch {
			return errors.New("you are already on the main branch; there is no branch to finish")
		}

		prompt := fmt.Sprintf("This will delete your local branch '%s' and switch to '%s'. Are you sure it has been merged?", currentBranch, mainBranch)
		confirmed, err := presenter.PromptForConfirmation(prompt)
		if err != nil {
			return err
		}
		if !confirmed {
			presenter.Info("Aborted by user.")
			return nil
		}

		presenter.Step("Switching to '%s'...", mainBranch)
		if err := gitClient.SwitchBranch(ctx, mainBranch); err != nil {
			return err
		}
		presenter.Step("Pulling latest changes for '%s'...", mainBranch)
		if err := gitClient.PullRebase(ctx, mainBranch); err != nil {
			presenter.Error("Failed to pull latest changes for main branch: %v", err)
			return err
		}
		presenter.Step("Deleting local branch '%s'...", currentBranch)
		if err := ExecClient.Execute(ctx, ".", "git", "branch", "-d", currentBranch); err != nil {
			presenter.Warning("Could not delete branch with '-d' (likely not fully merged). Trying '-D'...")
			if errForce := ExecClient.Execute(ctx, ".", "git", "branch", "-D", currentBranch); errForce != nil {
				presenter.Error("Failed to force delete branch '%s': %v", currentBranch, errForce)
				return errForce
			}
		}
		presenter.Success("Successfully cleaned up '%s' and updated '%s'.", currentBranch, mainBranch)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(gitTidyCmd)
	gitTidyCmd.AddCommand(gitTidyFinishCmd)
}
