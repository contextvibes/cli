// Package finish provides the command to finish a feature branch.
package finish

import (
	_ "embed"
	"errors"
	"fmt"
	//nolint:revive // Blank import for side effects (though none obvious here, keeping for safety).
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
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var FinishCmd = &cobra.Command{
	Use: "finish",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Finishing work on the current branch.")

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

		currentBranch, err := gitClient.GetCurrentBranchName(ctx)
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		if currentBranch == gitClient.MainBranchName() {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("cannot create a pull request from the main branch")
		}

		//nolint:noinlineerr // Inline check is standard.
		if err := gitClient.Push(ctx, currentBranch); err != nil {
			return fmt.Errorf("failed to push branch: %w", err)
		}

		if !globals.ExecClient.CommandExists("gh") {
			presenter.Warning("GitHub CLI ('gh') not found. Cannot create PR automatically.")

			return nil
		}

		confirmed, err := presenter.PromptForConfirmation("Create a GitHub Pull Request now?")
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			presenter.Info("Aborted. You can create the PR later by running 'gh pr create'.")

			return nil
		}

		presenter.Step("Running 'gh pr create'...")
		//nolint:noinlineerr // Inline check is standard.
		if err := globals.ExecClient.Execute(ctx, ".", "gh", "pr", "create", "--fill", "--web"); err != nil {
			return fmt.Errorf("gh pr create failed: %w", err)
		}

		presenter.Success("Pull Request created.")

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(finishLongDescription, nil)
	if err != nil {
		panic(err)
	}

	FinishCmd.Short = desc.Short
	FinishCmd.Long = desc.Long
}
