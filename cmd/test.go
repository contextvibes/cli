// cmd/test.go
package cmd

import (
	"context" // Ensure this is imported
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
	// "github.com/contextvibes/cli/internal/tools" // Should no longer be needed.
)

// Define an interface matching the methods used by the helpers below.
// This makes the helpers testable independently of the global ExecClient.
type execTestClientInterface interface {
	CommandExists(commandName string) bool
	Execute(ctx context.Context, dir string, commandName string, args ...string) error
	// CaptureOutput might be needed if specific test runners' output needs parsing (not used by current test helpers)
}

var testCmd = &cobra.Command{
	DisableFlagParsing: true,
	Use:   "test [args...]",
	Short: "Runs project-specific tests (e.g., go test, pytest).",
	Long: `Detects the project type (Go, Python) and runs the appropriate test command.
Any arguments passed to 'contextvibes test' will be forwarded to the underlying test runner.

- Go: Runs 'go test ./...'
- Python: Runs 'pytest' (if available). Falls back to 'python -m unittest discover' if pytest not found.

For other project types, or if no specific test runner is found, it will indicate no action.`,
	Example: `  contextvibes test
  contextvibes test -v  # Passes '-v' to 'go test' or 'pytest'
  contextvibes test tests/my_specific_test.py # Passes path to pytest`,
	// Args: cobra.ArbitraryArgs, // Keep commented out unless strictly needed and understood
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger    // From cmd/root.go
		if ExecClient == nil { // From cmd/root.go
			return errors.New("internal error: executor client not initialized")
		}
		if logger == nil {
			return errors.New("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr(), os.Stdin)
		ctx := context.Background() // Get context

		presenter.Summary("Running project tests.")

		cwd, err := os.Getwd()
		if err != nil {
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Test: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)

			return wrappedErr
		}

		presenter.Info("Detecting project type...")
		projType, err := project.Detect(cwd)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to detect project type: %w", err)
			logger.ErrorContext(ctx, "Test: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)

			return wrappedErr
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "test"), slog.String("type", string(projType)))

		var testErr error
		testExecuted := false

		switch projType {
		case project.Go:
			presenter.Header("Go Project Tests")
			// The codemod should have changed this call to include ExecClient
			testErr = executeGoTests(ctx, presenter, logger, ExecClient, cwd, args)
			testExecuted = true
		case project.Python:
			presenter.Header("Python Project Tests")
			// The codemod should have changed this call to include ExecClient
			testErr = executePythonTests(ctx, presenter, logger, ExecClient, cwd, args)
			testExecuted = true
		case project.Terraform, project.Pulumi:
			presenter.Info("Automated testing for %s projects is not yet implemented in this command.", projType)
			presenter.Advice("Consider using tools like Terratest or language-specific test frameworks manually.")
		case project.Unknown:
			presenter.Warning("Unknown project type. Cannot determine how to run tests.")
		default:
			presenter.Warning("No specific test execution logic for project type: %s", projType)
		}

		presenter.Newline()
		if !testExecuted && testErr == nil {
			presenter.Info("No tests were executed for the detected project type or configuration.")

			return nil
		}

		if testErr != nil {
			presenter.Error("Project tests failed.")
			logger.Error("Test command finished with errors", slog.String("source_command", "test"), slog.String("error", testErr.Error()))

			return testErr
		}

		presenter.Success("Project tests completed successfully.")
		logger.Info("Test command successful", slog.String("source_command", "test"))

		return nil
	},
}

// Manually updated signature: accepts ctx, execClient.
func executeGoTests(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execTestClientInterface, dir string, passThroughArgs []string) error {
	tool := "go"
	// This should have been updated by codemod
	if !execClient.CommandExists(tool) {
		errMsgForUser := fmt.Sprintf("'%s' command not found. Ensure Go is installed and in your PATH.", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Go test prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))

		return errors.New("go command not found")
	}

	testArgs := []string{"test", "./..."}
	testArgs = append(testArgs, passThroughArgs...)

	presenter.Info("Executing: %s %s", tool, strings.Join(testArgs, " "))
	logger.Info("Executing go test", slog.String("source_command", "test"), slog.String("tool", tool), slog.Any("args", testArgs))

	// This should have been updated by codemod
	err := execClient.Execute(ctx, dir, tool, testArgs...)
	if err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}

	return nil
}

// Manually updated signature: accepts ctx, execClient.
func executePythonTests(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execTestClientInterface, dir string, passThroughArgs []string) error {
	pytestTool := "pytest"
	pythonTool := "python"

	// This should have been updated by codemod
	if execClient.CommandExists(pytestTool) {
		presenter.Info("Executing: %s %s", pytestTool, strings.Join(passThroughArgs, " "))
		logger.Info("Executing pytest", slog.String("source_command", "test"), slog.String("tool", pytestTool), slog.Any("args", passThroughArgs))
		// This should have been updated by codemod
		err := execClient.Execute(ctx, dir, pytestTool, passThroughArgs...)
		if err != nil {
			return fmt.Errorf("pytest failed: %w", err)
		}

		return nil
	}

	presenter.Info("`pytest` not found. Attempting `python -m unittest discover`...")
	// This should have been updated by codemod
	if execClient.CommandExists(pythonTool) {
		unittestArgs := []string{"-m", "unittest", "discover"}

		presenter.Info("Executing: %s %s", pythonTool, strings.Join(unittestArgs, " "))
		logger.Info("Executing python unittest", slog.String("source_command", "test"), slog.String("tool", pythonTool), slog.Any("args", unittestArgs))
		// This should have been updated by codemod
		err := execClient.Execute(ctx, dir, pythonTool, unittestArgs...)
		if err != nil {
			return fmt.Errorf("python -m unittest discover failed: %w", err)
		}

		return nil
	}

	errMsgForUser := "Neither `pytest` nor `python` found. Cannot run Python tests."
	presenter.Error(errMsgForUser)
	logger.Error("Python test prerequisite failed", slog.String("reason", errMsgForUser))

	return errors.New("no python test runner found")
}

func init() {
	rootCmd.AddCommand(testCmd)
}
