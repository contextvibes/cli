// cmd/wrapup.go

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	// "os/exec" // No longer needed

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	// "github.com/contextvibes/cli/internal/tools" // No longer needed
	"github.com/spf13/cobra"
)

// assumeYes defined in root.go

var wrapupCmd = &cobra.Command{
	Use:   "wrapup",
	Short: "Finalizes daily work: stages, commits (default msg), pushes.",
	Long: `Performs the end-of-day workflow: checks for local changes or if the branch
is ahead of the remote, stages changes, commits with a standard message if needed,
and pushes the current branch.

Requires confirmation unless -y/--yes is specified.`,
	Example: `  contextvibes wrapup   # Checks state, stages, commits (if needed), pushes, after confirmation
  contextvibes wrapup -y # Performs wrapup without confirmation`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Starting end-of-day wrapup process.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Wrapup: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "wrapup"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "wrapup"))

		// --- Check Repository State ---
		presenter.Info("Checking repository state...")
		isClean, err := client.IsWorkingDirClean(ctx)
		if err != nil {
			presenter.Error("Failed checking working directory status: %v", err)
			return err // Client logs details
		}

		isAhead := false // Assume not ahead initially
		if isClean {
			presenter.Info("Working directory is clean.")
			// Only check if ahead if clean
			isAhead, err = client.IsBranchAhead(ctx)
			if err != nil {
				// Log the error but proceed; assume push might still be needed if we commit later
				logger.WarnContext(ctx, "Could not determine if branch is ahead", slog.String("source_command", "wrapup"), slog.String("error", err.Error()))
				presenter.Warning("Could not accurately determine push status: %v", err)
				// We will still propose push if a commit happens
			} else {
				if isAhead {
					presenter.Info("Local branch is ahead of remote.")
				} else {
					presenter.Info("Local branch is not ahead of remote.")
				}
			}
		} else {
			presenter.Info("Changes detected in working directory.")
			// If dirty, we'll commit, so push is definitely intended.
			isAhead = true // Treat as ahead for planning purposes if dirty
		}

		// --- Determine Actions ---
		actionCommit := !isClean              // Only commit if initially dirty
		actionPush := isAhead || actionCommit // Push if already ahead OR if we are about to commit

		commitMsg := "chore: Automated wrapup commit" // Default message

		if !actionCommit && !actionPush {
			presenter.Newline()
			presenter.Success("No actions needed (no local changes to commit and branch is not ahead).")
			logger.InfoContext(ctx, "Wrapup complete: no actions needed", slog.String("source_command", "wrapup"))
			// Add workflow advice even if no actions were needed (moved here for consistency)
			presenter.Newline()
			presenter.Advice("`wrapup` is an automated shortcut. For quality checks, custom commits, and pre-push sync, consider running:")
			presenter.Advice("  `contextvibes quality && contextvibes commit -m '...' && contextvibes sync`")
			return nil
		}

		// --- Confirmation ---
		presenter.Newline()
		presenter.Info("Proposed Wrapup Actions:")
		actionCounter := 1
		if actionCommit {
			presenter.Detail("%d. Stage all changes (git add .)", actionCounter)
			actionCounter++
			presenter.Detail("%d. Commit staged changes with message: '%s'", actionCounter, commitMsg)
			actionCounter++
		}
		if actionPush {
			presenter.Detail("%d. Push current branch to remote '%s'", actionCounter, client.RemoteName())
		}
		presenter.Newline()

		// *** MOVED ADVICE HERE ***
		presenter.Advice("`wrapup` is an automated shortcut. For quality checks, custom commits, and pre-push sync, consider running:")
		presenter.Advice("  `contextvibes quality && contextvibes commit -m '...' && contextvibes sync`")
		presenter.Newline() // Add space before the prompt
		// ***********************

		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(ctx, "Confirmation bypassed via flag", slog.String("source_command", "wrapup"), slog.Bool("yes_flag", true))
			confirmed = true
		} else {
			var promptErr error
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with wrapup?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading wrapup confirmation", slog.String("source_command", "wrapup"), slog.String("error", promptErr.Error()))
				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Wrapup aborted by user.")
			logger.InfoContext(ctx, "Wrapup aborted by user confirmation", slog.String("source_command", "wrapup"), slog.Bool("confirmed", false))
			return nil
		}
		logger.DebugContext(ctx, "Proceeding after wrapup confirmation", slog.String("source_command", "wrapup"), slog.Bool("confirmed", true))

		// --- Execute Actions ---
		presenter.Newline()
		commitActuallyHappened := false
		if actionCommit {
			presenter.Info("Staging all changes...")
			logger.DebugContext(ctx, "Attempting stage via client.AddAll", slog.String("source_command", "wrapup"))
			if err := client.AddAll(ctx); err != nil {
				presenter.Error("Failed to stage changes: %v", err)
				return err
			}

			// Check if staging actually resulted in changes to be committed
			logger.DebugContext(ctx, "Checking staged status after add", slog.String("source_command", "wrapup"))
			commitIsNeeded, err := client.HasStagedChanges(ctx)
			if err != nil {
				presenter.Error("Failed to check staged status after add: %v", err)
				return err
			}

			if commitIsNeeded {
				presenter.Info("Committing staged changes with message: '%s'...", commitMsg)
				logger.DebugContext(ctx, "Attempting commit via client.Commit", slog.String("source_command", "wrapup"))
				if err := client.Commit(ctx, commitMsg); err != nil {
					presenter.Error("Failed to commit changes: %v", err)
					return err
				}
				commitActuallyHappened = true
			} else {
				presenter.Info("No changes were staged after 'git add .', skipping commit step.")
				logger.InfoContext(ctx, "Commit skipped, no changes staged", slog.String("source_command", "wrapup"))
			}
		}

		if actionPush {
			// Determine branch to push (usually the current one)
			// GetCurrentBranchName is relatively safe even if called again
			branchToPush, branchErr := client.GetCurrentBranchName(ctx)
			if branchErr != nil {
				// If we can't get branch name now, it's a problem for push
				presenter.Error("Cannot determine current branch to push: %v", branchErr)
				return branchErr
			}

			presenter.Info("Pushing branch '%s' to remote '%s'...", branchToPush, client.RemoteName())
			logger.DebugContext(ctx, "Attempting push via client.Push", slog.String("source_command", "wrapup"), slog.String("branch", branchToPush))
			if err := client.Push(ctx, branchToPush); err != nil {
				// client.Push handles "up-to-date" logging internally and returns nil for it.
				// Only real errors should be returned here.
				presenter.Error("Failed to push changes: %v", err)
				return err
			}
			presenter.Info("Push successful or already up-to-date.")
		} else {
			presenter.Info("Skipping push step as no push action was planned.")
			logger.InfoContext(ctx, "Push skipped (not needed)", slog.String("source_command", "wrapup"))
		}

		// --- Final Status ---
		presenter.Newline()
		presenter.Success("Wrapup complete.")
		// Removed the advice from here as it's now shown before confirmation.
		// Just state what happened.
		if commitActuallyHappened && actionPush {
			presenter.Detail("Local changes were committed and the branch was pushed.")
		} else if commitActuallyHappened {
			presenter.Detail("Local changes were committed (push was not needed or skipped).")
		} else if actionPush {
			presenter.Detail("Branch was pushed (no local changes needed committing).")
		}

		logger.InfoContext(ctx, "Wrapup successful",
			slog.String("source_command", "wrapup"),
			slog.Bool("commit_executed", commitActuallyHappened),
			slog.Bool("push_executed", actionPush),
		)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(wrapupCmd)
}
