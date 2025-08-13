// cmd/deploy.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys infrastructure changes (terraform apply, pulumi up).",
	Long: `Detects the project type (Terraform, Pulumi), explains the deployment action,
and executes the deployment after confirmation (unless -y/--yes is specified).

- Terraform: Requires 'tfplan.out' from 'contextvibes plan'. Runs 'terraform apply tfplan.out'.
- Pulumi: Runs 'pulumi up', which internally includes a preview and confirmation.`,
	Example: `  # For Terraform:
  contextvibes plan    # First, generate the plan file (tfplan.out)
  contextvibes deploy  # Explain plan and prompt to apply tfplan.out
  contextvibes deploy -y # Apply tfplan.out without prompting

  # For Pulumi:
  contextvibes plan    # (Optional) Preview changes first
  contextvibes deploy  # Explain and run 'pulumi up' (includes preview & confirm)
  contextvibes deploy -y # Run 'pulumi up' without contextvibes confirmation`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		// Use global ExecClient from cmd/root.go
		if ExecClient == nil {
			return errors.New("internal error: executor client not initialized")
		}
		if logger == nil {
			return errors.New("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		ctx := context.Background()

		presenter.Summary("Attempting to deploy infrastructure changes.")

		cwd, err := os.Getwd()
		if err != nil {
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Deploy: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)

			return wrappedErr
		}

		presenter.Info("Detecting project type in %s...", presenter.Highlight(cwd))
		projType, err := project.Detect(cwd)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to detect project type: %w", err)
			logger.ErrorContext(
				ctx,
				"Deploy: Failed project detection",
				slog.String("error", err.Error()),
			)
			presenter.Error("Failed to detect project type: %v", err)

			return wrappedErr
		}

		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info(
			"Project detection result",
			slog.String("source_command", "deploy"),
			slog.String("type", string(projType)),
		)

		switch projType {
		case project.Terraform:
			// Pass ExecClient to the helper function
			return executeTerraformDeploy(ctx, presenter, logger, ExecClient, cwd, assumeYes)
		case project.Pulumi:
			// Pass ExecClient to the helper function
			return executePulumiDeploy(ctx, presenter, logger, ExecClient, cwd, assumeYes)
		case project.Go:
			presenter.Info("Deploy command is not applicable for Go projects.")

			return nil
		case project.Python:
			presenter.Info("Deploy command is not applicable for Python projects.")

			return nil
		case project.Unknown:
			errMsgForUser := "Unknown project type detected. Cannot determine deploy action."
			errMsgForError := "unknown project type detected"
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "deploy"))

			return errors.New(errMsgForError)
		default:
			errMsgForUser := fmt.Sprintf("Internal error: Unhandled project type '%s'", projType)
			errMsgForError := fmt.Sprintf("internal error: unhandled project type '%s'", projType)
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "deploy"))

			return errors.New(errMsgForError)
		}
	},
}

// Define an interface matching the methods used by the helpers below.
// This makes the helpers testable independently of the global ExecClient.
type execDeployClientInterface interface {
	CommandExists(commandName string) bool
	Execute(ctx context.Context, dir string, commandName string, args ...string) error
}

// executeTerraformDeploy now accepts execClient.
func executeTerraformDeploy(
	ctx context.Context,
	presenter *ui.Presenter,
	logger *slog.Logger,
	execClient execDeployClientInterface,
	dir string,
	skipConfirm bool,
) error {
	tool := "terraform"
	planFile := "tfplan.out"
	planFilePath := filepath.Join(dir, planFile)
	args := []string{"apply", "-auto-approve", planFile}

	if !execClient.CommandExists(tool) { // Use execClient
		errMsgForUser := fmt.Sprintf(
			"Command '%s' not found. Please ensure Terraform is installed and in your PATH.",
			tool,
		)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)

		presenter.Error(errMsgForUser)
		logger.Error(
			"Terraform deploy prerequisite failed",
			slog.String("reason", errMsgForUser),
			slog.String("tool", tool),
		)

		return errors.New(errMsgForError)
	}

	// Check for plan file using standard os.Stat - this doesn't involve executing a command
	logger.DebugContext(ctx, "Checking for Terraform plan file", slog.String("path", planFilePath))

	if _, err := os.Stat(planFilePath); os.IsNotExist(err) {
		errMsgForUser := fmt.Sprintf("Terraform plan file '%s' not found.", planFile)
		errMsgForError := "terraform plan file not found"

		presenter.Error(errMsgForUser)
		presenter.Advice("Please run `contextvibes plan` first to generate the plan file.")
		logger.Error(
			"Terraform deploy prerequisite failed: plan file missing",
			slog.String("plan_file", planFile),
		)

		return errors.New(errMsgForError)
	} else if err != nil {
		errMsgForUser := fmt.Sprintf("Error checking for plan file '%s': %v", planFilePath, err)
		errMsgForErrorBase := "error checking for plan file"

		presenter.Error(errMsgForUser)
		logger.Error("Terraform deploy: error stating plan file", slog.String("plan_file", planFilePath), slog.String("error", err.Error()))

		return fmt.Errorf("%s %s: %w", errMsgForErrorBase, planFilePath, err)
	}

	presenter.Info("Using Terraform plan file: %s", presenter.Highlight(planFile))

	// Confirmation logic remains the same
	presenter.Newline()
	presenter.Info("Proposed Deploy Action:")
	presenter.Detail("Apply the Terraform plan '%s' using command:", planFile)
	presenter.Detail("  %s %s", tool, strings.Join(args, " "))
	presenter.Newline()

	confirmed := false

	if skipConfirm {
		presenter.Info("Confirmation prompt bypassed via --yes flag.")
		logger.InfoContext(
			ctx,
			"Confirmation bypassed via flag",
			slog.String("source_command", "deploy"),
			slog.String("tool", tool),
			slog.Bool("yes_flag", true),
		)
		confirmed = true
	} else {
		var promptErr error

		confirmed, promptErr = presenter.PromptForConfirmation("Proceed with Terraform deployment?")
		if promptErr != nil {
			logger.ErrorContext(ctx, "Error reading deploy confirmation", slog.String("tool", tool), slog.String("error", promptErr.Error()))

			return promptErr
		}
	}

	if !confirmed {
		presenter.Info("Terraform deployment aborted by user.")
		logger.InfoContext(
			ctx,
			"Deploy aborted by user confirmation",
			slog.String("source_command", "deploy"),
			slog.String("tool", tool),
			slog.Bool("confirmed", false),
		)

		return nil
	}

	logger.DebugContext(
		ctx,
		"Proceeding after deploy confirmation",
		slog.String("source_command", "deploy"),
		slog.String("tool", tool),
		slog.Bool("confirmed", true),
	)

	// Execution using execClient
	presenter.Newline()
	presenter.Info("Starting Terraform apply...")
	logger.Info(
		"Executing terraform apply",
		slog.String("source_command", "deploy"),
		slog.String("tool", tool),
		slog.Any("args", args),
	)

	// Use execClient.Execute - terraform apply pipes its own output
	err := execClient.Execute(ctx, dir, tool, args...) // Use execClient
	if err != nil {
		// Error message from Execute should contain exit code info.
		// User will see the piped output from terraform apply itself.
		errMsgForUser := "'terraform apply' command failed."
		errMsgForError := "terraform apply command failed"

		presenter.Error(errMsgForUser)
		logger.Error(
			"Terraform apply command failed",
			slog.String("source_command", "deploy"),
			slog.String("error", err.Error()),
		)
		// Return a simpler error type, as the underlying error from Execute might not be needed by caller
		return errors.New(errMsgForError)
	}

	presenter.Newline()
	presenter.Success("Terraform apply successful.")
	logger.Info("Terraform apply successful", slog.String("source_command", "deploy"))

	return nil
}

// executePulumiDeploy now accepts execClient.
func executePulumiDeploy(
	ctx context.Context,
	presenter *ui.Presenter,
	logger *slog.Logger,
	execClient execDeployClientInterface,
	dir string,
	skipConfirm bool,
) error {
	tool := "pulumi"
	args := []string{"up"}

	if !execClient.CommandExists(tool) { // Use execClient
		errMsgForUser := fmt.Sprintf(
			"Command '%s' not found. Please ensure Pulumi is installed and in your PATH.",
			tool,
		)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)

		presenter.Error(errMsgForUser)
		logger.Error(
			"Pulumi deploy prerequisite failed",
			slog.String("reason", errMsgForUser),
			slog.String("tool", tool),
		)

		return errors.New(errMsgForError)
	}

	// Confirmation logic remains the same
	presenter.Newline()
	presenter.Info("Proposed Deploy Action:")
	presenter.Detail("Run '%s %s'.", tool, strings.Join(args, " "))
	presenter.Detail(
		"(Note: '%s up' will show its own preview and prompt for confirmation before making changes).",
		tool,
	)
	presenter.Newline()

	confirmed := false

	if skipConfirm {
		presenter.Info("Confirmation prompt (for contextvibes) bypassed via --yes flag.")
		logger.InfoContext(
			ctx,
			"Wrapper confirmation bypassed via flag",
			slog.String("source_command", "deploy"),
			slog.String("tool", tool),
			slog.Bool("yes_flag", true),
		)
		confirmed = true
	} else {
		var promptErr error

		confirmed, promptErr = presenter.PromptForConfirmation("Proceed to run 'pulumi up'?")
		if promptErr != nil {
			logger.ErrorContext(ctx, "Error reading deploy confirmation", slog.String("tool", tool), slog.String("error", promptErr.Error()))

			return promptErr
		}
	}

	if !confirmed {
		presenter.Info("'pulumi up' command aborted by user (before execution).")
		logger.InfoContext(
			ctx,
			"Deploy aborted by user confirmation",
			slog.String("source_command", "deploy"),
			slog.String("tool", tool),
			slog.Bool("confirmed", false),
		)

		return nil
	}

	logger.DebugContext(
		ctx,
		"Proceeding after deploy confirmation",
		slog.String("source_command", "deploy"),
		slog.String("tool", tool),
		slog.Bool("confirmed", true),
	)

	// Execution using execClient
	presenter.Newline()
	presenter.Info("Starting Pulumi execution ('%s %s')...", tool, strings.Join(args, " "))
	logger.Info(
		"Executing pulumi up",
		slog.String("source_command", "deploy"),
		slog.String("tool", tool),
		slog.Any("args", args),
	)

	// Use execClient.Execute - pulumi up pipes its own output
	err := execClient.Execute(ctx, dir, tool, args...) // Use execClient
	if err != nil {
		// Error message from Execute should contain exit code info.
		// User will see the piped output from pulumi up itself.
		errMsgForUser := "'pulumi up' command failed or was aborted by user during its execution."
		errMsgForError := "pulumi up command failed or aborted"

		presenter.Error(errMsgForUser)
		logger.Error(
			"Pulumi up command failed or aborted",
			slog.String("source_command", "deploy"),
			slog.String("error", err.Error()),
		)

		return errors.New(errMsgForError)
	}

	presenter.Newline()
	presenter.Success("Pulumi up completed successfully.")
	logger.Info("Pulumi up successful", slog.String("source_command", "deploy"))

	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
