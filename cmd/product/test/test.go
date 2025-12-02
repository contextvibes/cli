// Package test provides the command to run project tests.
package test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed test.md.tpl
var testLongDescription string

// TestCmd represents the test command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var TestCmd = &cobra.Command{
	DisableFlagParsing: true,
	Use:                "test [args...]",
	Example: `  contextvibes product test
  contextvibes product test -v  # Passes '-v' to 'go test' or 'pytest'
  contextvibes product test tests/my_specific_test.py # Passes path to pytest`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Running project tests.")

		cwd, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get current working directory: %v", err)

			return fmt.Errorf("failed to get working directory: %w", err)
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			presenter.Error("Failed to detect project type: %v", err)

			return fmt.Errorf("failed to detect project type: %w", err)
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))

		var testErr error
		testExecuted := false

		switch projType {
		case project.Go:
			presenter.Header("Go Project Tests")
			testErr = executeGoTests(ctx, presenter, globals.ExecClient, cwd, args)
			testExecuted = true
		case project.Python:
			presenter.Header("Python Project Tests")
			testErr = executePythonTests(ctx, presenter, globals.ExecClient, cwd, args)
			testExecuted = true
		case project.Terraform, project.Pulumi, project.Unknown:
			fallthrough
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

func executeGoTests(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
	dir string,
	passThroughArgs []string,
) error {
	tool := "go"
	testArgs := []string{"test", "./..."}
	testArgs = append(testArgs, passThroughArgs...)
	presenter.Info("Executing: %s %s", tool, strings.Join(testArgs, " "))

	err := execClient.Execute(ctx, dir, tool, testArgs...)
	if err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}

	return nil
}

func executePythonTests(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
	dir string,
	passThroughArgs []string,
) error {
	if execClient.CommandExists("pytest") {
		presenter.Info("Executing: pytest %s", strings.Join(passThroughArgs, " "))

		err := execClient.Execute(ctx, dir, "pytest", passThroughArgs...)
		if err != nil {
			return fmt.Errorf("pytest failed: %w", err)
		}

		return nil
	}

	presenter.Info("`pytest` not found. Attempting `python -m unittest discover`...")

	if execClient.CommandExists("python") {
		err := execClient.Execute(ctx, dir, "python", "-m", "unittest", "discover")
		if err != nil {
			return fmt.Errorf("python unittest failed: %w", err)
		}

		return nil
	}

	//nolint:err113 // Dynamic error is appropriate here.
	return errors.New("no python test runner found")
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(testLongDescription, nil)
	if err != nil {
		panic(err)
	}

	TestCmd.Short = desc.Short
	TestCmd.Long = desc.Long
}
