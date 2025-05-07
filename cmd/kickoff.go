// cmd/kickoff.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var branchNameFlag string

var kickoffCmd = &cobra.Command{
	Use:   "kickoff [--branch <branch-name>]",
	Short: "Prepares a new work branch. Requires clean main.",
	Long: `Performs the start-of-work workflow:
1. Requires a clean state on the main branch.
2. Updates the main branch from the remote (using configured default remote/main).
3. Creates and switches to a new branch. Name validation depends on .contextvibes.yaml.
   Default pattern if enabled: '^(feature|fix|docs|format)/.+'
4. Pushes the new branch to the remote and sets upstream tracking.

If --branch is not provided, you will be prompted.
Requires confirmation unless -y/--yes is specified.`,
	Example: `  contextvibes kickoff --branch feature/JIRA-123-new-widget
  contextvibes kickoff -b fix/login-bug -y
  contextvibes kickoff # Prompts for branch name`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		// Ensure LoadedAppConfig is available (should be initialized by root cmd's PersistentPreRunE)
		if LoadedAppConfig == nil {
			presenter.Error("Internal error: Application configuration not loaded.")
			logger.ErrorContext(ctx, "Kickoff failed: LoadedAppConfig is nil", slog.String("source_command", "kickoff"))
			return errors.New("application configuration not loaded")
		}

		presenter.Summary("Starting new work branch kickoff workflow.")

		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Kickoff: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}

		// Initialize GitClientConfig with values from LoadedAppConfig
		gitCfg := git.GitClientConfig{
			Logger:                logger,
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,     // From merged config
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch, // From merged config
			// GitExecutable could also be from config if made configurable in the future
		}
		logger.DebugContext(ctx, "Initializing GitClient with effective app config",
			slog.String("source_command", "kickoff"),
			slog.String("remote", gitCfg.DefaultRemoteName),
			slog.String("mainBranch", gitCfg.DefaultMainBranchName))

		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			// NewClient already logs details
			return err
		}

		mainBranchName := client.MainBranchName() // This will now be from app config via client
		remoteName := client.RemoteName()         // This will now be from app config via client
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
		} else {
			presenter.Info("Working directory is clean.")
		}

		if initialBranch != mainBranchName {
			preconditionsMet = false
			errMsg := fmt.Sprintf("Not on the main branch ('%s'). Current branch: '%s'.", mainBranchName, initialBranch)
			presenter.Error(errMsg)
			presenter.Advice("Switch to the main branch first using `git switch %s`.", mainBranchName)
			logger.WarnContext(ctx, "Kickoff prerequisite failed: not on main branch", slog.String("source_command", "kickoff"), slog.String("current_branch", initialBranch))
		} else {
			presenter.Info("Confirmed on main branch '%s'.", mainBranchName)
		}

		if !preconditionsMet {
			logger.ErrorContext(ctx, "Kickoff failed: prerequisites not met", slog.String("source_command", "kickoff"))
			return errors.New("prerequisites not met")
		}

		// --- Get Target Branch Name ---
		targetBranchName := strings.TrimSpace(branchNameFlag)
		if targetBranchName == "" {
			if assumeYes {
				errMsg := "Branch name is required via --branch flag when using --yes."
				presenter.Error(errMsg)
				logger.ErrorContext(ctx, "Kickoff failed: missing branch name with --yes", slog.String("source_command", "kickoff"))
				return errors.New("branch name required with --yes")
			}
			presenter.Newline()
			presenter.Info("Please provide the name for the new branch.")
			advicePattern := LoadedAppConfig.Validation.BranchName.Pattern
			if advicePattern == "" && (LoadedAppConfig.Validation.BranchName.Enable == nil || *LoadedAppConfig.Validation.BranchName.Enable) {
				// This case should ideally be caught by MergeWithDefaults ensuring a pattern is always there if enabled.
				// For safety, use the absolute default.
				advicePattern = "[default pattern, e.g., feature/name]"
			}
			presenter.Advice("Pattern (if validation enabled): %s", advicePattern)
			presenter.Advice("Examples: feature/JIRA-123-new-button, fix/auth-bug")
			for {
				targetBranchName, err = presenter.PromptForInput("New branch name: ")
				if err != nil {
					return err // Error reading input
				}
				targetBranchName = strings.TrimSpace(targetBranchName)
				if targetBranchName != "" {
					break
				}
				presenter.Warning("Branch name cannot be empty.")
			}
		}

		// --- Validate branch name based on configuration ---
		branchValidationRule := LoadedAppConfig.Validation.BranchName
		// Check if 'Enable' is explicitly set or if it's nil (meaning use default 'true')
		validationIsEnabled := branchValidationRule.Enable == nil || *branchValidationRule.Enable

		if validationIsEnabled {
			effectivePattern := branchValidationRule.Pattern
			if effectivePattern == "" {
				// This indicates a misconfiguration if validation is enabled but no pattern is set.
				// MergeWithDefaults in config package should prevent this by using the default pattern.
				presenter.Error("Internal Error: Branch validation is enabled but no pattern is defined in the effective configuration.")
				logger.ErrorContext(ctx, "Kickoff failed: branch validation enabled but pattern is empty", slog.String("source_command", "kickoff"))
				return errors.New("branch validation pattern misconfiguration")
			}

			branchNameRe, errRe := regexp.Compile(effectivePattern)
			if errRe != nil {
				presenter.Error("Internal error: Invalid branch name validation pattern ('%s') in configuration: %v", effectivePattern, errRe)
				logger.ErrorContext(ctx, "Kickoff failed: invalid configured branch regex", slog.String("source_command", "kickoff"), slog.String("pattern", effectivePattern), slog.String("error", errRe.Error()))
				return errors.New("invalid branch name regex in config")
			}
			if !branchNameRe.MatchString(targetBranchName) {
				errMsg := fmt.Sprintf("Invalid branch name: '%s'.", targetBranchName)
				presenter.Error(errMsg)
				presenter.Advice("Branch name must match the configured pattern: %s", effectivePattern)
				logger.ErrorContext(ctx, "Kickoff failed: invalid branch name format", slog.String("source_command", "kickoff"), slog.String("pattern", effectivePattern), slog.String("branch_name_provided", targetBranchName))
				return errors.New("invalid branch name format")
			}
			logger.DebugContext(ctx, "Branch name format validated successfully against pattern.", slog.String("source_command", "kickoff"), slog.String("pattern", effectivePattern), slog.String("branch_name", targetBranchName))
		} else {
			presenter.Info("Branch name validation is disabled by configuration.")
			logger.InfoContext(ctx, "Branch name validation skipped due to configuration", slog.String("source_command", "kickoff"))
		}
		logger.DebugContext(ctx, "Target branch name meets criteria (or validation disabled)", slog.String("source_command", "kickoff"), slog.String("name", targetBranchName))

		// --- Check if target branch already exists (locally) ---
		existsLocally, err := client.LocalBranchExists(ctx, targetBranchName)
		if err != nil {
			presenter.Error("Failed checking if target branch '%s' exists locally: %v", targetBranchName, err)
			return err
		}
		if existsLocally {
			errMsg := fmt.Sprintf("Branch '%s' already exists locally.", targetBranchName)
			presenter.Error(errMsg)
			presenter.Advice("Please choose a different name or delete/rename the existing local branch.")
			logger.ErrorContext(ctx, "Kickoff failed: branch already exists locally", slog.String("source_command", "kickoff"), slog.String("branch_name", targetBranchName))
			return errors.New("branch already exists locally")
		}

		// --- Plan Actions ---
		presenter.Newline()
		presenter.Info("Proposed Kickoff Actions:")
		presenter.Detail("1. Update main branch '%s' from remote '%s' (pull --rebase).", mainBranchName, remoteName)
		presenter.Detail("2. Create and switch to new local branch '%s' from '%s'.", targetBranchName, mainBranchName)
		presenter.Detail("3. Push new branch '%s' to '%s' and set upstream tracking.", targetBranchName, remoteName)
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
			return err
		}
		presenter.Info("Main branch update successful.")

		presenter.Newline()
		presenter.Info("Step 2: Creating and switching to new branch '%s' from '%s'...", targetBranchName, mainBranchName)
		if err := client.CreateAndSwitchBranch(ctx, targetBranchName, mainBranchName); err != nil {
			presenter.Error("Failed to create and switch to new branch '%s': %v", targetBranchName, err)
			return err
		}
		presenter.Info("Successfully created and switched to branch '%s'.", targetBranchName)

		presenter.Newline()
		presenter.Info("Step 3: Pushing new branch '%s' to '%s' and setting upstream...", targetBranchName, remoteName)
		if err := client.PushAndSetUpstream(ctx, targetBranchName); err != nil {
			presenter.Error("Failed to push new branch '%s': %v", targetBranchName, err)
			presenter.Advice("Check remote access and branch state. You may need to push manually using: git push --set-upstream %s %s", remoteName, targetBranchName)
			return err
		}
		presenter.Info("New branch '%s' pushed and upstream tracking set.", targetBranchName)

		// --- Final Verification ---
		finalBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Warning("Could not verify final current branch: %v", err)
		} else if finalBranch == targetBranchName {
			presenter.Newline()
			presenter.Success("Kickoff complete. You are now on new branch '%s'.", finalBranch)
			logger.InfoContext(ctx, "Kickoff successful", slog.String("source_command", "kickoff"), slog.String("final_branch", finalBranch))
		} else {
			errMsg := fmt.Sprintf("Workflow finished, but ended on branch '%s' (expected '%s'). Please check manually.", finalBranch, targetBranchName)
			presenter.Error(errMsg)
			logger.ErrorContext(ctx, "Kickoff finished on unexpected branch", slog.String("source_command", "kickoff"), slog.String("expected_branch", targetBranchName), slog.String("actual_branch", finalBranch))
			return errors.New(errMsg)
		}
		return nil
	},
}

func init() {
	kickoffCmd.Flags().StringVarP(&branchNameFlag, "branch", "b", "", "Name for the new branch (e.g., feature/JIRA-123-task-name)")
	rootCmd.AddCommand(kickoffCmd)
}
