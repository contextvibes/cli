// cmd/commit.go

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp" // Added for commit message validation
	"strings"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui" // Import the presenter package
	"github.com/spf13/cobra"
)

var commitMessageFlag string

// Regex for validating commit messages according to Conventional Commits.
// Format: <type>(<scope>): <subject>
// Valid types: BREAKING, feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
// Scope is optional. Subject is a concise description of the change.
const conventionalCommitRegexPattern = `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_]+\))?:\s.+`

var commitCmd = &cobra.Command{
	Use:   "commit -m <message>",
	Short: "Stages all changes and commits locally with a provided message.",
	Long: `Stages all current changes (tracked and untracked) in the working directory
and commits them locally using the message provided via the -m/--message flag.

The commit message should follow the Conventional Commits format:
  <type>(<scope>): <subject>
Examples:
  feat(login): add forgot password button
  fix(api): correct user data validation
  docs(readme): update installation instructions

Valid types: BREAKING, feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
Scope is optional. Subject is a concise description of the change.

Requires confirmation before committing unless -y/--yes is specified.
Does NOT automatically push.`,
	Example: `  contextvibes commit -m "feat(auth): Implement OTP login"
  contextvibes commit -m "fix: Correct typo in user model" -y`,
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
		if strings.TrimSpace(commitMessageFlag) == "" {
			errMsgForUser := "Commit message is required. Please provide one using the -m flag."
			errMsgForError := "commit message is required via -m flag"
			presenter.Error(errMsgForUser)
			presenter.Advice("Example: `%s commit -m \"feat(module): Your message\"`", cmd.CommandPath())
			logger.ErrorContext(ctx, "Commit failed: missing required message flag", slog.String("source_command", "commit"))
			return errors.New(errMsgForError)
		}
		finalCommitMessage := commitMessageFlag

		// --- Validate Commit Message Format ---
		commitMsgRe, _ := regexp.Compile(conventionalCommitRegexPattern) // Compile should not fail with this pattern
		if !commitMsgRe.MatchString(finalCommitMessage) {
			errMsgForUser := "Invalid commit message format."
			errMsgForError := "invalid commit message format"
			presenter.Error(errMsgForUser)
			presenter.Advice("Message should be: <type>(<scope>): <subject>")
			presenter.Detail("  Valid types: BREAKING, feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert")
			presenter.Detail("  Example: feat(ui): add new save button")
			presenter.Detail("  Your message: \"%s\"", finalCommitMessage)
			logger.ErrorContext(ctx, "Commit failed: invalid message format", slog.String("source_command", "commit"), slog.String("commit_message", finalCommitMessage))
			return errors.New(errMsgForError)
		}
		logger.DebugContext(ctx, "Commit message format validated successfully", slog.String("source_command", "commit"), slog.String("message", finalCommitMessage))

		// --- Summary ---
		presenter.Summary("Attempting to stage and commit changes locally.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get working directory: %v", err)
			logger.ErrorContext(ctx, "Commit: Failed getwd", slog.String("error", err.Error()))
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "commit"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed to initialize Git client: %v", err)
			return err // Client logs details
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "commit"))

		// --- Pre-Checks ---
		currentBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Error("Failed to get current branch name: %v", err)
			return err
		}
		if currentBranch == client.MainBranchName() {
			errMsg := fmt.Sprintf("Cannot commit directly on the main branch ('%s').", client.MainBranchName())
			presenter.Error(errMsg)
			presenter.Advice("Use `contextvibes kickoff` to start a new feature/fix branch first.")
			logger.ErrorContext(ctx, "Commit failed: attempt to commit on main branch", slog.String("source_command", "commit"), slog.String("branch", currentBranch))
			return errors.New("commit on main branch is disallowed by this command")
		}

		// --- Stage & Check ---
		logger.DebugContext(ctx, "Attempting to stage all changes (git add .)", slog.String("source_command", "commit"))
		if err := client.AddAll(ctx); err != nil {
			presenter.Error("Failed to stage changes: %v", err)
			return err // Client logs details
		}
		logger.DebugContext(ctx, "Staging completed.", slog.String("source_command", "commit"))

		hasStaged, err := client.HasStagedChanges(ctx)
		if err != nil {
			presenter.Error("Failed to check for staged changes: %v", err)
			return err // Client logs details
		}
		if !hasStaged {
			presenter.Info("No changes were staged for commit (working directory may have been clean or `git add .` staged nothing).")
			logger.InfoContext(ctx, "No staged changes found to commit.", slog.String("source_command", "commit"))
			return nil
		}

		// --- Fetch Git Status Details for Display ---
		statusOutput, _, statusErr := client.GetStatusShort(ctx)
		if statusErr != nil {
			logger.WarnContext(ctx, "Could not get short status for info block, proceeding with commit attempt.", slog.String("source_command", "commit"), slog.String("error", statusErr.Error()))
		}

		// --- Consolidated INFO Block ---
		presenter.Newline()
		presenter.InfoPrefixOnly() 

		fmt.Fprintf(presenter.Out(), "  Branch: %s\n", currentBranch)
		fmt.Fprintf(presenter.Out(), "  Commit Message:\n    \"%s\"\n", finalCommitMessage)
		fmt.Fprintf(presenter.Out(), "  Staged Changes:\n")
		if statusErr != nil {
			fmt.Fprintf(presenter.Out(), "    (Could not retrieve status details for display: %v)\n", statusErr)
		} else if strings.TrimSpace(statusOutput) == "" {
			fmt.Fprintln(presenter.Out(), "    (Staged changes detected, but `git status --short` was unexpectedly empty)")
		} else {
			scanner := bufio.NewScanner(strings.NewReader(statusOutput))
			for scanner.Scan() {
				fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
		}
		presenter.Newline()

		// --- Confirmation ---
		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(ctx, "Confirmation bypassed via --yes flag", slog.String("source_command", "commit"))
			confirmed = true
		} else {
			var promptErr error
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with this commit?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading confirmation for commit", slog.String("source_command", "commit"), slog.String("error", promptErr.Error()))
				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Commit aborted by user.")
			logger.InfoContext(ctx, "Commit aborted by user confirmation.", slog.String("source_command", "commit"))
			return nil
		}
		logger.DebugContext(ctx, "Proceeding with commit after confirmation.", slog.String("source_command", "commit"))

		// --- Execute Commit ---
		presenter.Info("Executing commit...")
		logger.DebugContext(ctx, "Attempting commit via GitClient.Commit", slog.String("source_command", "commit"))
		if err := client.Commit(ctx, finalCommitMessage); err != nil {
			presenter.Error("Commit command failed: %v", err)
			logger.ErrorContext(ctx, "client.Commit method failed", slog.String("source_command", "commit"), slog.String("error", err.Error()))
			return err
		}

		// --- Success & Advice ---
		presenter.Newline()
		presenter.Success("Commit created successfully locally.") 
		presenter.Advice("Consider syncing your changes using `contextvibes sync`.")
		logger.InfoContext(ctx, "Commit successful", slog.String("source_command", "commit"), slog.String("commit_message", finalCommitMessage))
		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessageFlag, "message", "m", "", "Commit message (required, format: type(scope): subject)")
	rootCmd.AddCommand(commitCmd)
}