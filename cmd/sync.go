// cmd/sync.go

package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"

	// "strings" // No longer needed directly here.

	"github.com/contextvibes/cli/internal/git" // Use GitClient
	"github.com/contextvibes/cli/internal/ui"  // Use Presenter

	// "github.com/contextvibes/cli/internal/tools" // No longer needed for Git/Prompts.
	"github.com/spf13/cobra"
)

// assumeYes defined in root.go

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs local branch with remote: ensures clean, pulls rebase, pushes if ahead.",
	Long: `Synchronizes the current local branch with its upstream remote counterpart.

Workflow:
1. Checks if the working directory is clean (no staged, unstaged, or untracked changes). Fails if dirty.
2. Determines current branch and remote.
3. Explains the plan (pull rebase, then push if needed).
4. Prompts for confirmation unless -y/--yes is specified.
5. Executes 'git pull --rebase'. Fails on conflicts or errors.
6. Checks if the local branch is ahead of the remote after the pull.
7. Executes 'git push' only if the branch was determined to be ahead.`,
	Example: `  contextvibes commit -m "Save work"  # Commit changes first if needed
  contextvibes sync                    # Sync the current branch
  contextvibes sync -y                 # Sync without confirmation`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return errors.New("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Synchronizing local branch with remote.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Sync: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)

			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger} // Use defaults for remote/main branch from config if needed later
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "sync"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)

			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "sync"))

		// --- Check Prerequisites ---
		presenter.Info("Checking working directory status...")
		isClean, err := client.IsWorkingDirClean(ctx)
		if err != nil {
			presenter.Error("Failed to check working directory status: %v", err)

			return err // Client logs details
		}
		if !isClean {
			errMsg := "Working directory has uncommitted changes (staged, unstaged, or untracked)."
			presenter.Error(errMsg)
			presenter.Advice("Please commit or stash your changes before syncing. Try `contextvibes commit -m \"...\"`.")
			logger.WarnContext(ctx, "Sync prerequisite failed: working directory not clean", slog.String("source_command", "sync"))

			return errors.New("working directory not clean") // Use specific error
		}
		presenter.Info("Working directory is clean.")

		currentBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			// Less critical if we can't get the name for display, but log it.
			presenter.Warning("Could not determine current branch name: %v", err)
			currentBranch = "current branch" // Use placeholder for messages
		}
		remoteName := client.RemoteName() // Get configured remote name

		// --- Confirmation ---
		presenter.Newline()
		presenter.Info("Proposed Sync Actions:")
		presenter.Detail("1. Update local branch '%s' from remote '%s' (git pull --rebase).", currentBranch, remoteName)
		presenter.Detail("2. Push local changes to remote '%s' if local branch is ahead after update (git push).", remoteName)
		presenter.Newline()

		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(ctx, "Confirmation bypassed via flag", slog.String("source_command", "sync"), slog.Bool("yes_flag", true))
			confirmed = true
		} else {
			var promptErr error
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with sync?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading sync confirmation", slog.String("source_command", "sync"), slog.String("error", promptErr.Error()))

				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Sync aborted by user.")
			logger.InfoContext(ctx, "Sync aborted by user confirmation", slog.String("source_command", "sync"), slog.Bool("confirmed", false))

			return nil
		}
		logger.DebugContext(ctx, "Proceeding after sync confirmation", slog.String("source_command", "sync"), slog.Bool("confirmed", true))

		// --- Execute Sync ---
		presenter.Newline()
		presenter.Info("Step 1: Updating local branch '%s' from '%s'...", currentBranch, remoteName)
		// Note: PullRebase uses runGit, which pipes output. User will see git's output directly.
		if err := client.PullRebase(ctx, currentBranch); err != nil {
			presenter.Error("Error during 'git pull --rebase'. Resolve conflicts manually and then run 'contextvibes sync' again if needed.")
			// Client logs details
			// Return specific error from PullRebase
			return err
		}
		presenter.Info("Pull --rebase successful.") // User info

		presenter.Newline()
		presenter.Info("Step 2: Checking if push is needed...")
		isAhead, err := client.IsBranchAhead(ctx)
		if err != nil {
			// This is more serious, as we can't determine push status
			presenter.Error("Failed to check if branch is ahead of remote: %v", err)
			// Client logs details
			return err
		}

		if !isAhead {
			presenter.Info("Local branch '%s' is not ahead of remote '%s'. Push is not required.", currentBranch, remoteName)
			logger.InfoContext(ctx, "Push not required after pull", slog.String("source_command", "sync"))
		} else {
			presenter.Info("Local branch '%s' is ahead of remote '%s'. Pushing changes...", currentBranch, remoteName)
			logger.DebugContext(ctx, "Attempting push via client.Push", slog.String("source_command", "sync"))
			// Note: Push uses runGit, which pipes output. User will see git's output.
			if err := client.Push(ctx, currentBranch); err != nil {
				// Push method handles "up-to-date" gracefully, only real errors are returned
				presenter.Error("Error during 'git push': %v", err)
				// Client logs details
				return err
			}
			presenter.Info("Push successful.") // User info
		}

		presenter.Newline()
		presenter.Success("Sync completed successfully.") // Use Success
		logger.InfoContext(ctx, "Sync successful", slog.String("source_command", "sync"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
