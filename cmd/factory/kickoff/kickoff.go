// cmd/factory/kickoff/kickoff.go
package kickoff

import (
	_ "embed"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

//go:embed kickoff.md.tpl
var kickoffLongDescription string

var branchNameFlag string

// KickoffCmd represents the kickoff command.
var KickoffCmd = &cobra.Command{
	Use:  "kickoff [--branch <branch-name>]",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		gitClient, err := git.NewClient(ctx, ".", git.GitClientConfig{
			Logger:                globals.AppLogger,
			DefaultRemoteName:     globals.LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: globals.LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              globals.ExecClient.UnderlyingExecutor(),
		})
		if err != nil {
			return err
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
			return err
		}

		runner := workflow.NewRunner(presenter, globals.AssumeYes)

		return runner.Run(
			ctx,
			"Daily Development Kickoff",
			&workflow.CheckOnMainBranchStep{GitClient: gitClient, Presenter: presenter},
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
