// cmd/factory/tidy/tidy.go
package tidy

import (
	_ "embed"
	"errors"
	"fmt"
	_ "os"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed tidy.md.tpl
var tidyLongDescription string

// TidyCmd represents the tidy command
var TidyCmd = &cobra.Command{
	Use: "tidy",
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("--- Finishing Merged Branch Workflow ---")

		gitClient, err := git.NewClient(
			ctx,
			".",
			git.GitClientConfig{
				Logger:   globals.AppLogger,
				Executor: globals.ExecClient.UnderlyingExecutor(),
			},
		)
		if err != nil {
			return err
		}

		mainBranch := gitClient.MainBranchName()
		currentBranch, err := gitClient.GetCurrentBranchName(ctx)
		if err != nil {
			return err
		}
		if currentBranch == mainBranch {
			return errors.New("you are already on the main branch; there is no branch to finish")
		}

		prompt := fmt.Sprintf(
			"This will delete your local branch '%s' and switch to '%s'. Are you sure it has been merged?",
			currentBranch,
			mainBranch,
		)
		confirmed, err := presenter.PromptForConfirmation(prompt)
		if err != nil {
			return err
		}
		if !confirmed {
			presenter.Info("Aborted by user.")
			return nil
		}

		if err := gitClient.SwitchBranch(ctx, mainBranch); err != nil {
			return err
		}
		if err := gitClient.PullRebase(ctx, mainBranch); err != nil {
			return err
		}

		if err := globals.ExecClient.Execute(ctx, ".", "git", "branch", "-d", currentBranch); err != nil {
			presenter.Warning(
				"Could not delete branch with '-d' (likely not fully merged). Trying '-D'...",
			)
			if errForce := globals.ExecClient.Execute(ctx, ".", "git", "branch", "-D", currentBranch); errForce != nil {
				return errForce
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

func init() {
	desc, err := cmddocs.ParseAndExecute(tidyLongDescription, nil)
	if err != nil {
		panic(err)
	}
	TidyCmd.Short = desc.Short
	TidyCmd.Long = desc.Long
}
