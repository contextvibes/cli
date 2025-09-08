// cmd/factory/plan/plan.go
package plan

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"os/exec"

	"github.com/contextvibes/cli/internal/cmddocs"
	internal_exec "github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed plan.md.tpl
var planLongDescription string

// PlanCmd represents the plan command
var PlanCmd = &cobra.Command{
	Use:     "plan",
	Example: `  contextvibes factory plan`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		logger, ok := cmd.Context().Value("logger").(*slog.Logger)
		if !ok { return errors.New("logger not found in context") }
		execClient, ok := cmd.Context().Value("execClient").(*internal_exec.ExecutorClient)
		if !ok { return errors.New("execClient not found in context") }
		ctx := cmd.Context()

		cwd, err := os.Getwd()
		if err != nil { return err }

		projType, err := project.Detect(cwd)
		if err != nil { return err }

		switch projType {
		case project.Terraform:
			return executeTerraformPlan(ctx, presenter, logger, execClient, cwd)
		case project.Pulumi:
			return executePulumiPreview(ctx, presenter, logger, execClient, cwd)
		default:
			presenter.Info("Plan command is not applicable for this project type.")
			return nil
		}
	},
}

func executeTerraformPlan(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient *internal_exec.ExecutorClient, dir string) error {
	_, _, err := execClient.CaptureOutput(ctx, dir, "terraform", "plan", "-out=tfplan.out")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
			presenter.Info("Terraform plan indicates changes are needed.")
			presenter.Advice("Plan saved to tfplan.out. Run `contextvibes factory deploy` to apply.")
			return nil
		}
		presenter.Error("'terraform plan' command failed.")
		return errors.New("terraform plan failed")
	}
	presenter.Info("Terraform plan successful (no changes detected).")
	return nil
}

func executePulumiPreview(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient *internal_exec.ExecutorClient, dir string) error {
	if err := execClient.Execute(ctx, dir, "pulumi", "preview"); err != nil {
		return errors.New("pulumi preview failed")
	}
	presenter.Success("Pulumi preview completed successfully.")
	return nil
}

func init() {
	desc, err := cmddocs.ParseAndExecute(planLongDescription, nil)
	if err != nil {
		panic(err)
	}
	PlanCmd.Short = desc.Short
	PlanCmd.Long = desc.Long
}
