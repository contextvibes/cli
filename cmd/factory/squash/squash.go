// Package squash provides the command to squash commits on a feature branch.
package squash

import (
	"context"
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

// NewSquashCmd creates and returns the squash command.
func NewSquashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "squash",
		Short: "Squashes all commits on the current feature branch into one.",
		Long: `Performs a "Soft Reset" to the merge-base of the main branch.
This stages all changes from your multiple commits into a single pending commit.
It automatically generates '_contextvibes.md' containing the diff, allowing
you to use an AI to generate the summary message before committing.`,
		Example:           `  contextvibes factory squash`,
		GroupID:           "factory",
		RunE:              runSquash,
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
}

func runSquash(cmd *cobra.Command, _ []string) error {
	presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
	ctx := cmd.Context()

	client, err := initializeGitClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize git client: %w", err)
	}

	state := &workflow.SquashState{}
	runner := workflow.NewRunner(presenter, globals.AssumeYes)

	steps := []workflow.Step{
		&workflow.EnsureCleanOrSaveStep{GitClient: client, Presenter: presenter, AssumeYes: globals.AssumeYes},
		&workflow.AnalyzeBranchStep{GitClient: client, Presenter: presenter, State: state},
		&workflow.SoftResetStep{GitClient: client, Presenter: presenter, State: state, AssumeYes: globals.AssumeYes},
		&workflow.GenerateSquashPromptStep{GitClient: client, Presenter: presenter, State: state},
		&workflow.CommitSquashStep{GitClient: client, Presenter: presenter, State: state, AssumeYes: globals.AssumeYes},
		&workflow.ForcePushStep{GitClient: client, Presenter: presenter, State: state, AssumeYes: globals.AssumeYes},
	}

	if err := runner.Run(ctx, "Squashing Feature Branch", steps...); err != nil {
		return fmt.Errorf("squash workflow failed: %w", err)
	}

	return nil
}

func initializeGitClient(ctx context.Context) (*git.GitClient, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	gitCfg := git.GitClientConfig{
		Logger:                globals.AppLogger,
		Executor:              globals.ExecClient.UnderlyingExecutor(),
		GitExecutable:         "git",
		DefaultRemoteName:     globals.LoadedAppConfig.Git.DefaultRemote,
		DefaultMainBranchName: globals.LoadedAppConfig.Git.DefaultMainBranch,
	}

	client, err := git.NewClient(ctx, cwd, gitCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create git client: %w", err)
	}

	return client, nil
}
