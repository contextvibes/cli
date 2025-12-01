// Package deploy provides the command to deploy infrastructure.
package deploy

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
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
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var DeployCmd = &cobra.Command{
	Use:     "deploy",
	Example: `  contextvibes factory deploy`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
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
		case project.Go, project.Python, project.Unknown:
			fallthrough
		default:
			presenter.Info("Deploy command is not applicable for this project type (%s).", projType)

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

	if _, err := os.Stat(planFilePath); err != nil {
		if os.IsNotExist(err) {
			presenter.Error("Terraform plan file '%s' not found.", planFile)
			presenter.Advice("Please run `contextvibes factory plan` first.")

			return errors.New("plan file not found")
		}

		return fmt.Errorf("failed to check plan file: %w", err)
	}

	presenter.Info("Proposed Deploy Action: Apply Terraform plan '%s'", planFile)

	if !skipConfirm {
		confirmed, err := presenter.PromptForConfirmation("Proceed with Terraform deployment?")
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirmed {
			presenter.Info("Deployment aborted.")

			return nil
		}
	}

	err := execClient.Execute(ctx, dir, "terraform", "apply", "-auto-approve", planFile)
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
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
		if err != nil {
			return fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirmed {
			presenter.Info("Deployment aborted.")

			return nil
		}
	}

	err := execClient.Execute(ctx, dir, "pulumi", "up")
	if err != nil {
		return fmt.Errorf("pulumi up failed: %w", err)
	}

	return nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(deployLongDescription, nil)
	if err != nil {
		panic(err)
	}

	DeployCmd.Short = desc.Short
	DeployCmd.Long = desc.Long
}
