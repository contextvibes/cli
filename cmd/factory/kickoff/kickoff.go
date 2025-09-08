// cmd/factory/kickoff/kickoff.go
package kickoff

import (
	_ "embed"
	"errors"
	"log/slog"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

//go:embed kickoff.md.tpl
var kickoffLongDescription string

var branchNameFlag string

// KickoffCmd represents the kickoff command
var KickoffCmd = &cobra.Command{
	Use:   "kickoff [--branch <branch-name>]",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		logger, ok := cmd.Context().Value("logger").(*slog.Logger)
		if !ok { return errors.New("logger not found in context") }
		execClient, ok := cmd.Context().Value("execClient").(*exec.ExecutorClient)
		if !ok { return errors.New("execClient not found in context") }
		loadedAppConfig, ok := cmd.Context().Value("config").(*config.Config)
		if !ok { return errors.New("config not found in context") }
		assumeYes, ok := cmd.Context().Value("assumeYes").(bool)
		if !ok { return errors.New("assumeYes not found in context") }
		ctx := cmd.Context()

		gitClient, err := git.NewClient(ctx, ".", git.GitClientConfig{
			Logger:                logger,
			DefaultRemoteName:     loadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: loadedAppConfig.Git.DefaultMainBranch,
			Executor:              execClient.UnderlyingExecutor(),
		})
		if err != nil {
			return err
		}

		validatedBranchName, err := workflow.GetValidatedBranchName(ctx, branchNameFlag, loadedAppConfig, presenter, gitClient, assumeYes)
		if err != nil {
			return err
		}

		runner := workflow.NewRunner(presenter, assumeYes)
		return runner.Run(
			ctx,
			"Daily Development Kickoff",
			&workflow.CheckOnMainBranchStep{GitClient: gitClient, Presenter: presenter},
			&workflow.CheckAndPromptStashStep{GitClient: gitClient, Presenter: presenter, AssumeYes: assumeYes},
			&workflow.UpdateMainBranchStep{GitClient: gitClient},
			&workflow.CreateAndPushBranchStep{GitClient: gitClient, BranchName: validatedBranchName},
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
	KickoffCmd.Flags().StringVarP(&branchNameFlag, "branch", "b", "", "Name for the new feature branch")
}
