// cmd/factory/deploy/deploy.go
package deploy

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed deploy.md.tpl
var deployLongDescription string

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:     "deploy",
	Example: `  contextvibes factory deploy`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			return err
		}

		switch projType {
		case project.Terraform:
			return executeTerraformDeploy(
				ctx,
				presenter,
				globals.ExecClient,
				cwd,
				globals.AssumeYes,
			)
		case project.Pulumi:
			return executePulumiDeploy(ctx, presenter, globals.ExecClient, cwd, globals.AssumeYes)
		default:
			presenter.Info("Deploy command is not applicable for this project type.")
			return nil
		}
	},
}

func executeTerraformDeploy(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
	dir string,
	skipConfirm bool,
) error {
	planFile := "tfplan.out"
	planFilePath := filepath.Join(dir, planFile)
	if _, err := os.Stat(planFilePath); os.IsNotExist(err) {
		presenter.Error("Terraform plan file '%s' not found.", planFile)
		presenter.Advice("Please run `contextvibes factory plan` first.")
		return errors.New("plan file not found")
	}

	presenter.Info("Proposed Deploy Action: Apply Terraform plan '%s'", planFile)
	if !skipConfirm {
		confirmed, err := presenter.PromptForConfirmation("Proceed with Terraform deployment?")
		if err != nil || !confirmed {
			presenter.Info("Deployment aborted.")
			return nil
		}
	}
	return execClient.Execute(ctx, dir, "terraform", "apply", "-auto-approve", planFile)
}

func executePulumiDeploy(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
	dir string,
	skipConfirm bool,
) error {
	presenter.Info("Proposed Deploy Action: Run 'pulumi up'")
	if !skipConfirm {
		confirmed, err := presenter.PromptForConfirmation("Proceed to run 'pulumi up'?")
		if err != nil || !confirmed {
			presenter.Info("Deployment aborted.")
			return nil
		}
	}
	return execClient.Execute(ctx, dir, "pulumi", "up")
}

func init() {
	desc, err := cmddocs.ParseAndExecute(deployLongDescription, nil)
	if err != nil {
		panic(err)
	}
	DeployCmd.Short = desc.Short
	DeployCmd.Long = desc.Long
}
