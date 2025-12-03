// Package message provides the command to generate commit message prompts.
package message

import (
	_ "embed"
	"fmt"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

//go:embed message.md.tpl
var messageLongDescription string

// MessageCmd represents the craft message command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var MessageCmd = &cobra.Command{
	Use:     "message",
	Aliases: []string{"commit", "msg"},
	Short:   "Generates a prompt for an AI to write your commit message.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		// 1. Initialize Git Client
		//nolint:exhaustruct // Partial config is sufficient.
		gitCfg := git.GitClientConfig{
			Logger:                globals.AppLogger,
			DefaultRemoteName:     globals.LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: globals.LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, ".", gitCfg)
		if err != nil {
			return fmt.Errorf("failed to initialize git client: %w", err)
		}

		// 2. Initialize Workflow Runner
		runner := workflow.NewRunner(presenter, globals.AssumeYes)

		// 3. Define and Run Steps
		return runner.Run(
			ctx,
			"Crafting Commit Message Prompt",
			&workflow.EnsureNotMainBranchStep{
				GitClient: client,
				Presenter: presenter,
			},
			&workflow.EnsureStagedStep{
				GitClient: client,
				Presenter: presenter,
				AssumeYes: globals.AssumeYes,
			},
			&workflow.GenerateCommitPromptStep{
				GitClient: client,
				Presenter: presenter,
			},
		)
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(messageLongDescription, nil)
	if err != nil {
		panic(err)
	}

	MessageCmd.Short = desc.Short
	MessageCmd.Long = desc.Long
}
