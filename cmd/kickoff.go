// cmd/kickoff.go

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/contextvibes/cli/internal/git" // Use GitClient
	"github.com/contextvibes/cli/internal/ui"  // Use Presenter
	// "github.com/contextvibes/cli/internal/tools" // No longer needed
	"github.com/spf13/cobra"
)

// assumeYes defined in root.go

// We might want these configurable via GitClientConfig later,
// but for now, access them via client methods if needed, or define constants here.
// For simplicity, let's get them from the initialized client.
// var gitRemoteName = "origin"
// var gitMainBranchName = "main"

var kickoffCmd = &cobra.Command{
	Use:   "kickoff",
	Short: "Prepares daily dev branch (dev-YYYY-MM-DD). Requires clean main.",
	Long: `Performs the start-of-day workflow: requires a clean state on the main branch,
updates main from the remote, then creates/switches to a 'dev-YYYY-MM-DD' branch,
and ensures it's pushed/up-to-date with the remote.

Requires confirmation unless -y/--yes is specified.`,
	Example: `  contextvibes kickoff    # Checks state, updates main, creates/updates dev branch, confirms
  contextvibes kickoff -y  # Does the same without confirmation`,
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

		presenter.Summary("Starting daily kickoff workflow.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Kickoff: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger} // Use defaults from config
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "kickoff"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "kickoff"))

		// --- Get Configured Names ---
		mainBranchName := client.MainBranchName()
		remoteName := client.RemoteName()
		presenter.Info("Using remote '%s' and main branch '%s'.", remoteName, mainBranchName)

		// --- Check Prerequisites ---
		presenter.Info("Checking prerequisites...")
		initialBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Error("Failed to get current branch: %v", err)
			return err
		}
		logger.DebugContext(ctx, "Initial branch", slog.String("branch", initialBranch), slog.String("source_command", "kickoff"))

		isClean, err := client.IsWorkingDirClean(ctx)
		if err != nil {
			presenter.Error("Failed checking working directory status: %v", err)
			return err
		}

		preconditionsMet := true
		if !isClean {
			preconditionsMet = false
			errMsg := "Working directory is not clean. Kickoff requires a clean state."
			presenter.Error(errMsg)
			presenter.Advice("Commit or stash changes first. Try `contextvibes commit -m \"...\"` or `git stash`.")
			logger.WarnContext(ctx, "Kickoff prerequisite failed: working directory not clean", slog.String("source_command", "kickoff"))
			// No need to check branch if already failed
		} else {
			presenter.Info("Working directory is clean.")
			// Only check branch if clean
			if initialBranch != mainBranchName {
				preconditionsMet = false
				errMsg := fmt.Sprintf("Not on the main branch ('%s'). Current branch: '%s'.", mainBranchName, initialBranch)
				presenter.Error(errMsg)
				presenter.Advice("Switch to the main branch first using `git switch %s`.", mainBranchName)
				logger.WarnContext(ctx, "Kickoff prerequisite failed: not on main branch", slog.String("source_command", "kickoff"), slog.String("current_branch", initialBranch))
			} else {
				presenter.Info("Confirmed on main branch '%s'.", mainBranchName)
			}
		}

		if !preconditionsMet {
			logger.ErrorContext(ctx, "Kickoff failed: prerequisites not met", slog.String("source_command", "kickoff"))
			return errors.New("prerequisites not met") // Simple error message
		}

		// --- Plan Actions ---
		presenter.Info("Prerequisites met.")
		dailyBranchName := "dev-" + time.Now().Format("2006-01-02")
		logger.DebugContext(ctx, "Determined daily branch name", slog.String("name", dailyBranchName), slog.String("source_command", "kickoff"))

		existsLocally, err := client.LocalBranchExists(ctx, dailyBranchName)
		if err != nil {
			presenter.Error("Failed checking if daily branch '%s' exists: %v", dailyBranchName, err)
			return err
		}

		presenter.Newline()
		presenter.Info("Proposed Kickoff Actions:")
		presenter.Detail("1. Update main branch '%s' from remote '%s' (pull --rebase).", mainBranchName, remoteName)
		if existsLocally {
			presenter.Detail("2. Switch to existing local branch '%s'.", dailyBranchName)
			presenter.Detail("3. Attempt to update '%s' from '%s' (pull --rebase).", dailyBranchName, remoteName)
		} else {
			presenter.Detail("2. Create and switch to new branch '%s' from '%s'.", dailyBranchName, mainBranchName)
			presenter.Detail("3. Push new branch '%s' to '%s' and set upstream.", dailyBranchName, remoteName)
		}
		presenter.Newline()

		// --- Confirmation ---
		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(ctx, "Confirmation bypassed via flag", slog.String("source_command", "kickoff"), slog.Bool("yes_flag", true))
			confirmed = true
		} else {
			var promptErr error
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with kickoff workflow?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading kickoff confirmation", slog.String("source_command", "kickoff"), slog.String("error", promptErr.Error()))
				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Kickoff aborted by user.")
			logger.InfoContext(ctx, "Kickoff aborted by user confirmation", slog.String("source_command", "kickoff"), slog.Bool("confirmed", false))
			return nil
		}
		logger.DebugContext(ctx, "Proceeding after kickoff confirmation", slog.String("source_command", "kickoff"), slog.Bool("confirmed", true))

		// --- Execute Actions ---
		presenter.Newline()
		presenter.Info("Step 1: Updating main branch '%s'...", mainBranchName)
		if err := client.PullRebase(ctx, mainBranchName); err != nil {
			presenter.Error("Failed to update main branch '%s'. Resolve conflicts or issues and retry.", mainBranchName)
			// Client logs details
			return err
		}
		presenter.Info("Main branch update successful.")

		presenter.Newline()
		presenter.Info("Step 2 & 3: Preparing daily branch '%s'...", dailyBranchName)
		if existsLocally {
			presenter.Info("Switching to existing branch '%s'...", dailyBranchName)
			if err := client.SwitchBranch(ctx, dailyBranchName); err != nil {
				presenter.Error("Failed to switch to existing branch '%s': %v", dailyBranchName, err)
				return err
			}
			presenter.Info("Attempting update for '%s' from '%s'...", dailyBranchName, remoteName)
			// PullRebase might fail if branch doesn't exist remotely yet or conflicts exist
			if err := client.PullRebase(ctx, dailyBranchName); err != nil {
				// This is often not fatal, just means branch isn't on remote or needs manual pull
				presenter.Warning("Could not automatically pull rebase for '%s': %v", dailyBranchName, err)
				presenter.Advice("Manual pull may be needed if remote branch exists and has changes.")
				logger.WarnContext(ctx, "Non-critical failure during pull rebase for existing daily branch", slog.String("source_command", "kickoff"), slog.String("branch", dailyBranchName), slog.String("error", err.Error()))
			} else {
				presenter.Info("Branch '%s' is up-to-date with remote.", dailyBranchName)
			}
		} else {
			presenter.Info("Creating and switching to new branch '%s' from '%s'...", dailyBranchName, mainBranchName)
			if err := client.CreateAndSwitchBranch(ctx, dailyBranchName, mainBranchName); err != nil {
				presenter.Error("Failed to create new branch '%s': %v", dailyBranchName, err)
				return err
			}
			presenter.Info("Pushing new branch '%s' to '%s' and setting upstream...", dailyBranchName, remoteName)
			if err := client.PushAndSetUpstream(ctx, dailyBranchName); err != nil {
				presenter.Error("Failed to push new branch '%s': %v", dailyBranchName, err)
				presenter.Advice("Check remote access and branch state. You may need to push manually.")
				return err
			}
			presenter.Info("New branch pushed and tracking set.")
		}

		// --- Final Verification ---
		finalBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Warning("Could not verify final current branch: %v", err)
		} else if finalBranch == dailyBranchName {
			presenter.Newline()
			presenter.Success("Daily kickoff complete. You are now on branch '%s'.", finalBranch)
			logger.InfoContext(ctx, "Kickoff successful", slog.String("source_command", "kickoff"), slog.String("final_branch", finalBranch))
		} else {
			// This shouldn't happen if switch commands succeeded
			errMsg := fmt.Sprintf("Workflow finished, but ended on branch '%s' (expected '%s'). Please check manually.", finalBranch, dailyBranchName)
			presenter.Error(errMsg)
			logger.ErrorContext(ctx, "Kickoff finished on unexpected branch", slog.String("source_command", "kickoff"), slog.String("expected_branch", dailyBranchName), slog.String("actual_branch", finalBranch))
			return errors.New(errMsg)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(kickoffCmd)
}
