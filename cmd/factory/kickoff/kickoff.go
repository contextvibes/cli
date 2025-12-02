// Package kickoff provides the command to start a new task or project.
package kickoff

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

//go:embed kickoff.md.tpl
var kickoffLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var branchNameFlag string

// KickoffCmd represents the kickoff command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var KickoffCmd = &cobra.Command{
	Use:  "kickoff [--branch <branch-name>]",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		//nolint:exhaustruct // Partial config is sufficient.
		gitClient, err := git.NewClient(ctx, ".", git.GitClientConfig{
			Logger:                globals.AppLogger,
			DefaultRemoteName:     globals.LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: globals.LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              globals.ExecClient.UnderlyingExecutor(),
		})
		if err != nil {
			return fmt.Errorf("failed to initialize git client: %w", err)
		}

		validatedBranchName, err := workflow.GetValidatedBranchName(
			ctx,
			branchNameFlag,
			globals.LoadedAppConfig,
			presenter,
			gitClient,
			globals.AssumeYes,
		)
		if err != nil {
			return fmt.Errorf("branch validation failed: %w", err)
		}

		runner := workflow.NewRunner(presenter, globals.AssumeYes)

		return runner.Run(
			ctx,
			"Daily Development Kickoff",
			&workflow.CheckOnMainBranchStep{GitClient: gitClient, Presenter: presenter},
			//nolint:exhaustruct // DidStash is an output field.
			&workflow.CheckAndPromptStashStep{
				GitClient: gitClient,
				Presenter: presenter,
				AssumeYes: globals.AssumeYes,
			},
			&workflow.UpdateMainBranchStep{GitClient: gitClient},
			&workflow.CreateAndPushBranchStep{
				GitClient:  gitClient,
				BranchName: validatedBranchName,
			},
		)
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(kickoffLongDescription, nil)
	if err != nil {
		panic(err)
	}

	KickoffCmd.Short = desc.Short
	KickoffCmd.Long = desc.Long
	KickoffCmd.Flags().
		StringVarP(&branchNameFlag, "branch", "b", "", "Name for the new feature branch")
}
