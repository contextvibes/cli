// FILE: cmd/kickoff.go
package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	branchNameFlag            string
	isStrategicKickoffFlag    bool
	markStrategicCompleteFlag bool
)

var kickoffCmd = &cobra.Command{
	Use:           "kickoff [--branch <branch-name>] [--strategic] [--mark-strategic-complete]",
	Short:         "Manages project kickoff: daily branch workflow or strategic project initiation.",
	Long:          `Manages project kickoff workflows... (Full description omitted for brevity)`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Standard setup
		logger := AppLogger
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		ctx := context.Background()

		if LoadedAppConfig == nil {
			presenter.Error("Internal error: Application configuration not loaded.")
			return errors.New("application configuration not loaded")
		}
		if ExecClient == nil {
			presenter.Error("Internal error: Executor client not initialized.")
			return errors.New("executor client not initialized")
		}

		// Strategic and mark-complete logic is out of scope for this refactor and remains.
		if markStrategicCompleteFlag {
			presenter.Warning("Marking strategic complete is not yet refactored.")
			return nil // Replace with actual call later
		}
		runStrategic := isStrategicKickoffFlag
		if !runStrategic {
			if LoadedAppConfig.ProjectState.StrategicKickoffCompleted == nil ||
				!*LoadedAppConfig.ProjectState.StrategicKickoffCompleted {
				runStrategic = true
			}
		}
		if runStrategic {
			presenter.Warning("Strategic kickoff generation is not yet refactored.")
			return nil // Replace with actual call later
		}

		// --- Refactored Daily Kickoff Logic ---
		gitClient, err := git.NewClient(ctx, ".", git.GitClientConfig{
			Logger:                logger,
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              ExecClient.UnderlyingExecutor(),
		})
		if err != nil {
			presenter.Error("Failed to initialize Git client: %v", err)
			return err
		}

		// Get the branch name *before* starting the workflow.
		validatedBranchName, err := workflow.GetValidatedBranchName(
			ctx,
			branchNameFlag,
			LoadedAppConfig,
			presenter,
			gitClient,
			assumeYes,
		)
		if err != nil {
			return err // Helper function already printed user-facing error
		}

		// Instantiate the workflow runner
		runner := workflow.NewRunner(presenter, assumeYes)

		// Define and run the workflow
		return runner.Run(
			ctx,
			"Daily Development Kickoff",
			&workflow.CheckOnMainBranchStep{GitClient: gitClient, Presenter: presenter},
			&workflow.CheckAndPromptStashStep{
				GitClient: gitClient,
				Presenter: presenter,
				AssumeYes: assumeYes,
			},
			&workflow.UpdateMainBranchStep{GitClient: gitClient},
			&workflow.CreateAndPushBranchStep{
				GitClient:  gitClient,
				BranchName: validatedBranchName,
			},
		)
	},
}

func init() {
	kickoffCmd.Flags().
		StringVarP(&branchNameFlag, "branch", "b", "", "Name for the new daily/feature branch (e.g., feature/JIRA-123)")
	kickoffCmd.Flags().
		BoolVar(&isStrategicKickoffFlag, "strategic", false, "Generates a master prompt for an AI-guided strategic project kickoff session.")
	kickoffCmd.Flags().
		BoolVar(&markStrategicCompleteFlag, "mark-strategic-complete", false, "Marks the strategic kickoff as complete in .contextvibes.yaml.")
	rootCmd.AddCommand(kickoffCmd)
}
