// cmd/factory/finish/finish.go
package finish

import (
	_ "embed"
	"errors"
	_ "os"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed finish.md.tpl
var finishLongDescription string

// FinishCmd represents the finish command.
var FinishCmd = &cobra.Command{
	Use: "finish",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Finishing work on the current branch.")

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

		currentBranch, err := gitClient.GetCurrentBranchName(ctx)
		if err != nil {
			return err
		}
		if currentBranch == gitClient.MainBranchName() {
			return errors.New("cannot create a pull request from the main branch")
		}

		if err := gitClient.Push(ctx, currentBranch); err != nil {
			return err
		}

		if !globals.ExecClient.CommandExists("gh") {
			presenter.Warning("GitHub CLI ('gh') not found. Cannot create PR automatically.")

			return nil
		}

		confirmed, err := presenter.PromptForConfirmation("Create a GitHub Pull Request now?")
		if err != nil {
			return err
		}
		if !confirmed {
			presenter.Info("Aborted. You can create the PR later by running 'gh pr create'.")

			return nil
		}

		presenter.Step("Running 'gh pr create'...")
		if err := globals.ExecClient.Execute(ctx, ".", "gh", "pr", "create", "--fill", "--web"); err != nil {
			return err
		}

		presenter.Success("Pull Request created.")

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(finishLongDescription, nil)
	if err != nil {
		panic(err)
	}

	FinishCmd.Short = desc.Short
	FinishCmd.Long = desc.Long
}
