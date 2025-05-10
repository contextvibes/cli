// cmd/kickoff.go
package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/kickoff"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	branchNameFlag            string 
	isStrategicKickoffFlag  bool   
	markStrategicCompleteFlag bool   
)

var kickoffCmd = &cobra.Command{
	Use:   "kickoff [--branch <branch-name>] [--strategic] [--mark-strategic-complete]",
	Short: "Manages project kickoff: daily branch workflow or strategic project initiation.",
	Long: `Manages project kickoff workflows.

Default Behavior (Daily Kickoff, if strategic completed):
  - Requires a clean state on the main branch.
  - Updates the main branch, creates a new daily/feature branch, and pushes it.
  - Uses --branch flag or prompts for name (respects .contextvibes.yaml validation).

Strategic Kickoff Prompt Generation (--strategic, or if first run):
  - Initiates a brief interactive session to gather basic project details.
  - Generates a comprehensive master prompt file (STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md).
  - User takes this prompt to an external AI to complete the detailed strategic kickoff.

Marking Strategic Kickoff as Complete (--mark-strategic-complete):
  - Updates '.contextvibes.yaml' to indicate the strategic kickoff has been done.
  - This enables the daily kickoff workflow for subsequent runs without '--strategic'.

Global --yes flag (from root command) bypasses confirmations for daily kickoff actions.`,
	Example: `  # Daily Kickoff Examples (assumes strategic kickoff was previously marked complete)
  contextvibes kickoff --branch feature/new-login
  contextvibes kickoff -b fix/bug-123 -y
  contextvibes kickoff # Prompts for branch name

  # Strategic Kickoff Prompt Generation
  contextvibes kickoff --strategic 
  contextvibes kickoff             # Runs strategic prompt generation if first time

  # Mark Strategic Kickoff as Done (after user completes session with external AI)
  contextvibes kickoff --mark-strategic-complete`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		if LoadedAppConfig == nil {
			presenter.Error("Internal error: Application configuration not loaded.")
			logger.ErrorContext(ctx, "Kickoff cmd failed: LoadedAppConfig is nil", slog.String("source_command", "kickoff"))
			return errors.New("application configuration not loaded")
		}
		if ExecClient == nil {
			presenter.Error("Internal error: Executor client not initialized.")
			logger.ErrorContext(ctx, "Kickoff cmd failed: ExecClient is nil", slog.String("source_command", "kickoff"))
			return errors.New("executor client not initialized")
		}

		var configFilePath string
		// Attempt to find .contextvibes.yaml in repo root
		repoCfgPath, findPathErr := config.FindRepoRootConfigPath(ExecClient) 
		if findPathErr != nil { // Error finding repo root (e.g., not a git repo)
			logger.WarnContext(ctx, "Could not determine git repository root. '.contextvibes.yaml' will be looked for/created in current directory.",
				slog.String("source_command", "kickoff"), slog.Any("find_path_error", findPathErr))
			cwd, _ := os.Getwd()
			configFilePath = filepath.Join(cwd, config.DefaultConfigFileName)
		} else if repoCfgPath == "" { // Repo root found, but .contextvibes.yaml doesn't exist there
			logger.InfoContext(ctx, "'.contextvibes.yaml' not found in repository root. It will be created there if needed.",
				slog.String("source_command", "kickoff"))
			// Get repo root again to ensure correct path for creation
			repoRootForCreation, _, _ := ExecClient.CaptureOutput(context.Background(), ".", "git", "rev-parse", "--show-toplevel")
			cleanRoot := strings.TrimSpace(repoRootForCreation)
			if cleanRoot == "" || cleanRoot == "." { // Fallback if somehow rev-parse fails here
				cwd, _ := os.Getwd()
				cleanRoot = cwd
			}
			configFilePath = filepath.Join(cleanRoot, config.DefaultConfigFileName)
		} else { // Config file found in repo root
			configFilePath = repoCfgPath
		}
		logger.DebugContext(ctx, "Determined config file path for kickoff operations", 
			slog.String("path", configFilePath), slog.String("source_command", "kickoff"))


		workDir, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get working directory: %v", err)
			logger.ErrorContext(ctx, "Kickoff cmd: Failed getwd", slog.String("error", err.Error()))
			return err
		}
		
		var gitClt *git.GitClient
		gitClientConfig := git.GitClientConfig{
			Logger:                logger, 
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              ExecClient.UnderlyingExecutor(), 
		}
		// Try to initialize Git client. It's okay if it fails for some strategic operations.
		gitClt, err = git.NewClient(ctx, workDir, gitClientConfig)
		if err != nil {
			// Only log a warning here. The orchestrator will decide if a nil gitClient is fatal for the chosen operation.
			logger.WarnContext(ctx, "Kickoff cmd: Git client initialization failed. Some operations might be limited.", 
				slog.String("source_command", "kickoff"), 
				slog.String("error", err.Error()))
			// gitClt will be nil
		}

		// The global 'assumeYes' is set by rootCmd.PersistentPreRunE or by flag parsing.
		orchestrator := kickoff.NewOrchestrator(logger, LoadedAppConfig, presenter, gitClt, configFilePath)

		if markStrategicCompleteFlag {
			// User wants to mark the strategic kickoff as done.
			if isStrategicKickoffFlag || branchNameFlag != "" {
				presenter.Warning("--mark-strategic-complete is mutually exclusive with --strategic and --branch. Ignoring other flags.")
				logger.WarnContext(ctx, "Redundant flags with --mark-strategic-complete", 
					slog.Bool("strategic_flag", isStrategicKickoffFlag),
					slog.String("branch_flag", branchNameFlag))
			}
			err = orchestrator.MarkStrategicKickoffComplete(ctx)
		} else {
			// Pass the global 'assumeYes' (from cmd/root.go, bound to rootCmd.PersistentFlags())
			err = orchestrator.ExecuteKickoff(ctx, isStrategicKickoffFlag, branchNameFlag, assumeYes)
		}
		
		if err != nil {
			// Orchestrator's presenter methods should have already shown user-friendly messages.
			// Log the error from the orchestrator for the AI trace log.
			logger.ErrorContext(ctx, "Kickoff command execution resulted in error", 
				slog.String("source_command", "kickoff"), 
				slog.Any("error", err)) // err already includes context from orchestrator
			// Return error to signify failure to Cobra, avoids double-printing if presenter handled it.
			return err 
		}

		logger.InfoContext(ctx, "Kickoff command completed successfully.", slog.String("source_command", "kickoff"))
		return nil
	},
}

func init() {
	kickoffCmd.Flags().StringVarP(&branchNameFlag, "branch", "b", "", "Name for the new daily/feature branch (e.g., feature/JIRA-123)")
	kickoffCmd.Flags().BoolVar(&isStrategicKickoffFlag, "strategic", false, "Generates a master prompt for an AI-guided strategic project kickoff session.")
	kickoffCmd.Flags().BoolVar(&markStrategicCompleteFlag, "mark-strategic-complete", false, "Marks the strategic kickoff as complete in .contextvibes.yaml.")
	
	rootCmd.AddCommand(kickoffCmd)
}
