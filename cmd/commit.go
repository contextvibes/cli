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

	"github.com/contextvibes/cli/internal/config" // Import for DefaultCommitMessagePattern
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui" // Import the presenter package
	"github.com/spf13/cobra"
)

var commitMessageFlag string

// The hardcoded conventionalCommitRegexPattern is no longer the primary source.
// It will be determined by configuration or fallback to config.DefaultCommitMessagePattern.
// const conventionalCommitRegexPattern = `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_]+\))?:\s.+`

var commitCmd = &cobra.Command{
	Use:   "commit -m <message>",
	Short: "Stages all changes and commits locally with a provided message.",
	Long: `Stages all current changes (tracked and untracked) in the working directory
and commits them locally using the message provided via the -m/--message flag.

Commit message validation is active by default, expecting a Conventional Commits format.
This can be configured (pattern or disabled) in '.contextvibes.yaml'.
Default pattern if validation is enabled and no custom pattern is set:
  ` + config.DefaultCommitMessagePattern + `
Example (using default Conventional Commits):
  feat(login): add forgot password button
  fix(api): correct user data validation

Requires confirmation before committing unless -y/--yes is specified.
Does NOT automatically push.`,
	Example: `  contextvibes commit -m "feat(auth): Implement OTP login"
  contextvibes commit -m "fix: Correct typo in user model" -y
  contextvibes commit -m "My custom message" # (if validation is disabled or pattern allows)
  contextvibes commit --config-validation-pattern="^TASK-[0-9]+: .+" -m "TASK-123: Implement feature" # (Example of custom pattern if it were a flag, actual via .contextvibes.yaml)`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			_, _ = fmt.Fprintln(os.Stderr, "[ERROR] Internal error: logger not initialized")
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		ctx := context.Background()

		if LoadedAppConfig == nil {
			presenter.Error("Internal error: Application configuration not loaded.")
			logger.ErrorContext(
				ctx,
				"Commit failed: LoadedAppConfig is nil",
				slog.String("source_command", "commit"),
			)
			return errors.New("application configuration not loaded")
		}

		// --- Validate Input ---
		if strings.TrimSpace(commitMessageFlag) == "" {
			errMsgForUser := "Commit message is required. Please provide one using the -m flag."
			errMsgForError := "commit message is required via -m flag"
			presenter.Error(errMsgForUser)
			presenter.Advice(
				"Example: `%s commit -m \"feat(module): Your message\"`",
				cmd.CommandPath(),
			)
			logger.ErrorContext(
				ctx,
				"Commit failed: missing required message flag",
				slog.String("source_command", "commit"),
			)
			return errors.New(errMsgForError)
		}
		finalCommitMessage := commitMessageFlag

		// --- Validate Commit Message Format Based on Configuration ---
		commitMsgValidationRule := LoadedAppConfig.Validation.CommitMessage
		validationIsEnabled := commitMsgValidationRule.Enable == nil ||
			*commitMsgValidationRule.Enable // Default to true if nil

		effectivePattern := ""
		patternSource := ""

		if validationIsEnabled {
			logger.InfoContext(
				ctx,
				"Commit message validation is enabled.",
				slog.String("source_command", "commit"),
			)
			effectivePattern = commitMsgValidationRule.Pattern
			patternSource = "from .contextvibes.yaml"

			if effectivePattern == "" {
				effectivePattern = config.DefaultCommitMessagePattern // Fallback to built-in default
				patternSource = "default built-in"
				logger.DebugContext(
					ctx,
					"Using default built-in commit message pattern because configured pattern is empty.",
					slog.String("source_command", "commit"),
					slog.String("pattern", effectivePattern),
				)
			} else {
				logger.DebugContext(ctx, "Using commit message pattern from configuration.", slog.String("source_command", "commit"), slog.String("pattern", effectivePattern))
			}

			if effectivePattern == "" {
				// This should not happen if DefaultCommitMessagePattern is always defined
				presenter.Error(
					"Internal Error: Commit message validation is enabled but no pattern is defined or defaulted.",
				)
				logger.ErrorContext(
					ctx,
					"Commit failed: validation enabled but no pattern available",
					slog.String("source_command", "commit"),
				)
				return errors.New(
					"commit validation pattern misconfiguration (empty effective pattern)",
				)
			}

			commitMsgRe, compileErr := regexp.Compile(effectivePattern)
			if compileErr != nil {
				errMsgForUser := fmt.Sprintf(
					"Internal error: Invalid commit message validation pattern ('%s') from %s.",
					effectivePattern,
					patternSource,
				)
				errMsgForError := "invalid commit message validation regex"
				presenter.Error(errMsgForUser)
				presenter.Advice("Error details: %v", compileErr)
				presenter.Advice(
					"Please check your .contextvibes.yaml or report this issue if using the default pattern.",
				)
				logger.ErrorContext(ctx, "Commit failed: invalid regex for commit message",
					slog.String("source_command", "commit"),
					slog.String("pattern", effectivePattern),
					slog.String("pattern_source", patternSource),
					slog.String("error", compileErr.Error()))
				return errors.New(errMsgForError)
			}

			if !commitMsgRe.MatchString(finalCommitMessage) {
				errMsgForUser := "Invalid commit message format."
				errMsgForError := "invalid commit message format"
				presenter.Error(errMsgForUser)
				presenter.Advice(
					"Message should match the pattern (%s): `%s`",
					patternSource,
					effectivePattern,
				)
				if patternSource != "default built-in" &&
					effectivePattern == config.DefaultCommitMessagePattern {
					presenter.Detail(
						" (Note: Configured pattern seems to be the same as the default Conventional Commits pattern.)",
					)
				} else if effectivePattern == config.DefaultCommitMessagePattern {
					presenter.Detail("  Default pattern expects: <type>(<scope>): <subject>")
					presenter.Detail("  Valid types: BREAKING, feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert")
					presenter.Detail("  Example: feat(ui): add new save button")
				}
				presenter.Detail("  Your message: \"%s\"", finalCommitMessage)
				logger.ErrorContext(ctx, "Commit failed: invalid message format",
					slog.String("source_command", "commit"),
					slog.String("commit_message", finalCommitMessage),
					slog.String("pattern", effectivePattern),
					slog.String("pattern_source", patternSource))
				return errors.New(errMsgForError)
			}
			logger.DebugContext(ctx, "Commit message format validated successfully",
				slog.String("source_command", "commit"),
				slog.String("message", finalCommitMessage),
				slog.String("pattern", effectivePattern),
				slog.String("pattern_source", patternSource))
		} else {
			presenter.Info("Commit message validation is disabled by configuration (.contextvibes.yaml).")
			logger.InfoContext(ctx, "Commit message validation skipped due to configuration", slog.String("source_command", "commit"))
		}

		// --- Summary ---
		presenter.Summary("Attempting to stage and commit changes locally.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get working directory: %v", err)
			logger.ErrorContext(ctx, "Commit: Failed getwd", slog.String("error", err.Error()))
			return err
		}
		// Pass relevant config values to GitClientConfig
		gitCfg := git.GitClientConfig{
			Logger:                logger,
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch,
		}
		logger.DebugContext(ctx, "Initializing GitClient with effective app config",
			slog.String("source_command", "commit"),
			slog.String("remote", gitCfg.DefaultRemoteName),
			slog.String("mainBranch", gitCfg.DefaultMainBranchName))
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
		// Use client.MainBranchName() which gets the effective main branch from its config
		if currentBranch == client.MainBranchName() {
			errMsg := fmt.Sprintf(
				"Cannot commit directly on the main branch ('%s').",
				client.MainBranchName(),
			)
			presenter.Error(errMsg)
			presenter.Advice("Use `contextvibes kickoff` to start a new feature/fix branch first.")
			logger.ErrorContext(
				ctx,
				"Commit failed: attempt to commit on main branch",
				slog.String("source_command", "commit"),
				slog.String("branch", currentBranch),
			)
			return errors.New("commit on main branch is disallowed by this command")
		}

		// --- Stage & Check ---
		logger.DebugContext(
			ctx,
			"Attempting to stage all changes (git add .)",
			slog.String("source_command", "commit"),
		)
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
			presenter.Info(
				"No changes were staged for commit (working directory may have been clean or `git add .` staged nothing).",
			)
			logger.InfoContext(
				ctx,
				"No staged changes found to commit.",
				slog.String("source_command", "commit"),
			)
			return nil
		}

		// --- Fetch Git Status Details for Display ---
		statusOutput, _, statusErr := client.GetStatusShort(ctx)
		if statusErr != nil {
			logger.WarnContext(
				ctx,
				"Could not get short status for info block, proceeding with commit attempt.",
				slog.String("source_command", "commit"),
				slog.String("error", statusErr.Error()),
			)
		}

		// --- Consolidated INFO Block ---
		presenter.Newline()
		presenter.InfoPrefixOnly()

		_, _ = fmt.Fprintf(presenter.Out(), "  Branch: %s\n", currentBranch)
		_, _ = fmt.Fprintf(presenter.Out(), "  Commit Message:\n    \"%s\"\n", finalCommitMessage)
		if validationIsEnabled {
			_, _ = fmt.Fprintf(
				presenter.Out(),
				"  Validation Pattern (%s):\n    `%s`\n",
				patternSource,
				effectivePattern,
			)
		} else {
			_, _ = fmt.Fprintln(presenter.Out(), "  Validation: Disabled by configuration")
		}
		_, _ = fmt.Fprintf(presenter.Out(), "  Staged Changes:\n")

		if statusErr != nil {
			_, _ = fmt.Fprintf(
				presenter.Out(),
				"    (Could not retrieve status details for display: %v)\n",
				statusErr,
			)
		} else if strings.TrimSpace(statusOutput) == "" {
			_, _ = fmt.Fprintln(presenter.Out(), "    (Staged changes detected, but `git status --short` was unexpectedly empty)")
		} else {
			scanner := bufio.NewScanner(strings.NewReader(statusOutput))
			for scanner.Scan() {
				_, _ = fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
		}
		presenter.Newline()

		// --- Confirmation ---
		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(
				ctx,
				"Confirmation bypassed via --yes flag",
				slog.String("source_command", "commit"),
			)
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
			logger.InfoContext(
				ctx,
				"Commit aborted by user confirmation.",
				slog.String("source_command", "commit"),
			)
			return nil
		}
		logger.DebugContext(
			ctx,
			"Proceeding with commit after confirmation.",
			slog.String("source_command", "commit"),
		)

		// --- Execute Commit ---
		presenter.Info("Executing commit...")
		logger.DebugContext(
			ctx,
			"Attempting commit via GitClient.Commit",
			slog.String("source_command", "commit"),
		)
		if err := client.Commit(ctx, finalCommitMessage); err != nil {
			presenter.Error("Commit command failed: %v", err)
			logger.ErrorContext(
				ctx,
				"client.Commit method failed",
				slog.String("source_command", "commit"),
				slog.String("error", err.Error()),
			)
			return err
		}

		// --- Success & Advice ---
		presenter.Newline()
		presenter.Success("Commit created successfully locally.")
		presenter.Advice("Consider syncing your changes using `contextvibes sync`.")
		logger.InfoContext(
			ctx,
			"Commit successful",
			slog.String("source_command", "commit"),
			slog.String("commit_message", finalCommitMessage),
		)
		return nil
	},
}

func init() {
	commitCmd.Flags().
		StringVarP(&commitMessageFlag, "message", "m", "", "Commit message (required)")
	// Long description of commitCmd already describes the default pattern and configurability
	rootCmd.AddCommand(commitCmd)
}
