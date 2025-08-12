// FILE: cmd/kickoff.go
package cmd

import (
	"context"
	"errors"
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
		logger := AppLogger
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		if LoadedAppConfig == nil {
			presenter.Error("Internal error: Application configuration not loaded.")
			return errors.New("application configuration not loaded")
		}
		if ExecClient == nil {
			presenter.Error("Internal error: Executor client not initialized.")
			return errors.New("executor client not initialized")
		}

		var configFilePath string
		repoCfgPath, _ := config.FindRepoRootConfigPath(ExecClient)
		if repoCfgPath == "" {
			repoRootForCreation, _, _ := ExecClient.CaptureOutput(
				context.Background(),
				".",
				"git",
				"rev-parse",
				"--show-toplevel",
			)
			cleanRoot := strings.TrimSpace(repoRootForCreation)
			if cleanRoot == "" || cleanRoot == "." {
				cwd, _ := os.Getwd()
				cleanRoot = cwd
			}
			configFilePath = filepath.Join(cleanRoot, config.DefaultConfigFileName)
		} else {
			configFilePath = repoCfgPath
		}

		workDir, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get working directory: %v", err)
			return err
		}

		var gitClt *git.GitClient
		gitClientConfig := git.GitClientConfig{
			Logger:                logger,
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              ExecClient.UnderlyingExecutor(),
		}
		gitClt, _ = git.NewClient(ctx, workDir, gitClientConfig)

		orchestrator := kickoff.NewOrchestrator(
			logger,
			LoadedAppConfig,
			presenter,
			gitClt,
			configFilePath,
			assumeYes,
		)

		if markStrategicCompleteFlag {
			return orchestrator.MarkStrategicKickoffComplete(ctx)
		}
		return orchestrator.ExecuteKickoff(ctx, isStrategicKickoffFlag, branchNameFlag)
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
