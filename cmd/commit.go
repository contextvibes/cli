// cmd/commit.go

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui" // Import the presenter package
	"github.com/spf13/cobra"
)

var commitMessageFlag string

// Assume AppLogger is initialized in rootCmd for AI file logging
// var AppLogger *slog.Logger // Defined in root.go

var commitCmd = &cobra.Command{
	Use:   "commit -m <message>",
	Short: "Stages all changes and commits locally with a provided message.",
	Long: `Stages all current changes (tracked and untracked) in the working directory
and commits them locally using the message provided via the -m/--message flag.

Requires confirmation before committing unless -y/--yes is specified.
Does NOT automatically push.`,
	Example: `  contextvibes commit -m "feat: Add new login feature" # Stages, confirms, commits
  contextvibes commit -m "Update README" -y            # Stages, skips confirmation, commits`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			fmt.Fprintln(os.Stderr, "[ERROR] Internal error: logger not initialized")
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		// --- Validate Input ---
		// After
		if strings.TrimSpace(commitMessageFlag) == "" {
			errMsgForUser := "Commit message is required. Please provide one using the -m flag."
			errMsgForError := "commit message is required via -m flag" // Lowercase, no punctuation
			presenter.Error(errMsgForUser)
			presenter.Advice("Example: `%s commit -m \"Your message\"`", cmd.CommandPath())
			logger.ErrorContext(ctx, "Commit failed: missing required message flag", slog.String("source_command", "commit"))
			return errors.New(errMsgForError) // <-- Use lowercase error value
		}
		finalCommitMessage := commitMessageFlag

		// --- Summary ---
		presenter.Summary("Attempting to stage and commit changes locally.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil { /* handle */
			presenter.Error("Failed getwd: %v", err)
			logger.ErrorContext(ctx, "", slog.Any("err", err))
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient" /*...*/)
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized" /*...*/)

		// --- Pre-Checks ---
		currentBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Error("Failed get branch: %v", err)
			return err
		}
		if currentBranch == client.MainBranchName() { /* handle */
			presenter.Error("Cannot commit on main")
			presenter.Advice("Use kickoff")
			logger.ErrorContext(ctx, "", slog.String("err", "commit on main"))
			return errors.New("commit on main")
		}
		// Don't print branch info separately here yet

		// --- Stage & Check ---
		logger.DebugContext(ctx, "Attempting stage via client.AddAll", slog.String("source_command", "commit"))
		if err := client.AddAll(ctx); err != nil {
			presenter.Error("Failed stage: %v", err)
			return err
		}
		logger.DebugContext(ctx, "client.AddAll completed", slog.String("source_command", "commit"))

		hasStaged, err := client.HasStagedChanges(ctx)
		if err != nil {
			presenter.Error("Failed check staged: %v", err)
			return err
		}
		if !hasStaged {
			presenter.Info("No changes were staged for commit (working directory may have been clean).")
			logger.InfoContext(ctx, "No staged changes found to commit.", slog.String("source_command", "commit"))
			return nil
		}
		// Don't print staged info separately here yet

		// --- Fetch Git Status Details ---
		statusOutput, _, statusErr := client.GetStatusShort(ctx)
		// Log error getting status, but proceed if possible as it's informational
		if statusErr != nil {
			logger.WarnContext(ctx, "Could not get short status for info block", slog.String("source_command", "commit"), slog.String("error", statusErr.Error()))
		}

		// --- Consolidated INFO Block ---
		presenter.Newline()
		// Print the INFO: prefix once using the presenter's color
		presenter.InfoPrefixOnly() // Assumes we add this helper to Presenter

		// Print indented lines using standard fmt to the presenter's output stream
		fmt.Fprintf(presenter.Out(), "  Branch: %s\n", currentBranch)
		fmt.Fprintf(presenter.Out(), "  Commit Message:\n    \"%s\"\n", finalCommitMessage)
		fmt.Fprintf(presenter.Out(), "  Staged Changes:\n")
		if statusErr != nil {
			fmt.Fprintf(presenter.Out(), "    (Could not retrieve status details: %v)\n", statusErr)
		} else if strings.TrimSpace(statusOutput) == "" {
			// Should be unlikely if HasStagedChanges was true, but handle anyway
			fmt.Fprintln(presenter.Out(), "    (None detected by status - unusual)")
		} else {
			// Indent each line of the status output
			scanner := bufio.NewScanner(strings.NewReader(statusOutput))
			for scanner.Scan() {
				fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
		}
		presenter.Newline() // Space after the info block

		// --- Confirmation ---
		// The proposed action is now clear from the INFO block above

		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.") // Use regular Info here
			logger.InfoContext(ctx, "Confirmation bypassed via flag" /* ... */)
			confirmed = true
		} else {
			var promptErr error
			// Prompt is now more direct as info was already presented
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with this commit?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading confirmation" /*...*/)
				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Commit aborted.") // Use Info
			logger.InfoContext(ctx, "Commit aborted by user confirmation." /* ... */)
			return nil
		}
		logger.DebugContext(ctx, "Proceeding after confirmation." /* ... */)

		// --- Execute Commit ---
		presenter.Info("Executing commit...") // Standard execution message
		logger.DebugContext(ctx, "Attempting commit via GitClient" /* ... */)
		if err := client.Commit(ctx, finalCommitMessage); err != nil {
			presenter.Error("Commit command failed: %v", err)
			return err
		}

		// --- Success & Advice ---
		presenter.Newline()
		presenter.Info("Commit successful.")
		presenter.Advice("Commit created locally. Consider syncing using `contextvibes sync`.")
		logger.InfoContext(ctx, "Commit successful" /* ... */)
		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessageFlag, "message", "m", "", "Commit message (required)")
	rootCmd.AddCommand(commitCmd)
}
