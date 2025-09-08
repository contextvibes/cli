// cmd/factory/sync/sync.go
package sync

import (
	_ "embed"
	"errors"
	"os"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed sync.md.tpl
var syncLongDescription string

// SyncCmd represents the sync command
var SyncCmd = &cobra.Command{
	Use:     "sync",
	Example: `  contextvibes factory sync`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Synchronizing local branch with remote.")

		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		gitCfg := git.GitClientConfig{
			Logger:   globals.AppLogger,
			Executor: globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}

		isClean, err := client.IsWorkingDirClean(ctx)
		if err != nil {
			return err
		}
		if !isClean {
			presenter.Error("Working directory has uncommitted changes.")
			presenter.Advice("Please commit or stash your changes before syncing.")
			return errors.New("working directory not clean")
		}

		currentBranch, _ := client.GetCurrentBranchName(ctx)
		presenter.Newline()
		presenter.Info("Proposed Sync Actions:")
		presenter.Detail("1. Update local branch '%s' from remote (git pull --rebase).", currentBranch)
		presenter.Detail("2. Push local changes to remote if ahead (git push).")
		presenter.Newline()

		if !globals.AssumeYes {
			confirmed, err := presenter.PromptForConfirmation("Proceed with sync?")
			if err != nil {
				return err
			}
			if !confirmed {
				presenter.Info("Sync aborted by user.")
				return nil
			}
		}

		if err := client.PullRebase(ctx, currentBranch); err != nil {
			presenter.Error("Error during 'git pull --rebase'. Resolve conflicts manually.")
			return err
		}

		isAhead, err := client.IsBranchAhead(ctx)
		if err != nil {
			return err
		}
		if isAhead {
			if err := client.Push(ctx, currentBranch); err != nil {
				return err
			}
		}

		presenter.Newline()
		presenter.Success("Sync completed successfully.")
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(syncLongDescription, nil)
	if err != nil {
		panic(err)
	}
	SyncCmd.Short = desc.Short
	SyncCmd.Long = desc.Long
}
