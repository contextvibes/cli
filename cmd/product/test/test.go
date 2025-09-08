// cmd/product/test/test.go
package test

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed test.md.tpl
var testLongDescription string

// TestCmd represents the test command
var TestCmd = &cobra.Command{
	DisableFlagParsing: true,
	Use:                "test [args...]",
	Example: `  contextvibes product test
  contextvibes product test -v  # Passes '-v' to 'go test' or 'pytest'
  contextvibes product test tests/my_specific_test.py # Passes path to pytest`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

		logger, ok := cmd.Context().Value("logger").(*slog.Logger)
		if !ok { return errors.New("logger not found in context") }
		execClient, ok := cmd.Context().Value("execClient").(*exec.ExecutorClient)
		if !ok { return errors.New("execClient not found in context") }
		ctx := cmd.Context()

		presenter.Summary("Running project tests.")

		cwd, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get current working directory: %v", err)
			return err
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			presenter.Error("Failed to detect project type: %v", err)
			return err
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))

		var testErr error
		testExecuted := false

		switch projType {
		case project.Go:
			presenter.Header("Go Project Tests")
			testErr = executeGoTests(ctx, presenter, logger, execClient, cwd, args)
			testExecuted = true
		case project.Python:
			presenter.Header("Python Project Tests")
			testErr = executePythonTests(ctx, presenter, logger, execClient, cwd, args)
			testExecuted = true
		default:
			presenter.Info("No specific test execution logic for project type: %s", projType)
		}

		presenter.Newline()
		if !testExecuted && testErr == nil {
			presenter.Info("No tests were executed.")
			return nil
		}

		if testErr != nil {
			presenter.Error("Project tests failed.")
			return testErr
		}

		presenter.Success("Project tests completed successfully.")
		return nil
	},
}

func executeGoTests(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient *exec.ExecutorClient, dir string, passThroughArgs []string) error {
	tool := "go"
	testArgs := []string{"test", "./..."}
	testArgs = append(testArgs, passThroughArgs...)
	presenter.Info("Executing: %s %s", tool, strings.Join(testArgs, " "))
	return execClient.Execute(ctx, dir, tool, testArgs...)
}

func executePythonTests(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient *exec.ExecutorClient, dir string, passThroughArgs []string) error {
	if execClient.CommandExists("pytest") {
		presenter.Info("Executing: pytest %s", strings.Join(passThroughArgs, " "))
		return execClient.Execute(ctx, dir, "pytest", passThroughArgs...)
	}
	presenter.Info("`pytest` not found. Attempting `python -m unittest discover`...")
	if execClient.CommandExists("python") {
		return execClient.Execute(ctx, dir, "python", "-m", "unittest", "discover")
	}
	return errors.New("no python test runner found")
}

func init() {
	desc, err := cmddocs.ParseAndExecute(testLongDescription, nil)
	if err != nil {
		panic(err)
	}
	TestCmd.Short = desc.Short
	TestCmd.Long = desc.Long
}
