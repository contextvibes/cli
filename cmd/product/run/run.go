// Package run provides the command to execute example applications.
package run

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed run.md.tpl
var runLongDescription string

// RunCmd represents the run command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var RunCmd = &cobra.Command{
	Use:     "run",
	Example: `  contextvibes product run`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Header("--- Example Runner ---")

		examples, err := findRunnableExamples(".")
		if err != nil {
			return err
		}
		if len(examples) == 0 {
			presenter.Warning("No runnable examples found in the './examples' directory.")

			return nil
		}

		choice, err := presenter.PromptForSelect("Please select an example to run:", examples)
		if err != nil || choice == "" {
			return nil // User aborted
		}

		if err := runVerificationChecks(ctx, presenter, globals.ExecClient, globals.LoadedAppConfig, choice); err != nil {
			return errors.New("prerequisite verification failed")
		}

		presenter.Newline()
		presenter.Step("Executing example: %s...", presenter.Highlight(choice))
		err = globals.ExecClient.Execute(ctx, ".", "go", "run", "./"+choice)
		if err != nil {
			return errors.New("example execution failed")
		}

		globals.AppLogger.InfoContext(ctx, "Successfully launched example", "example_path", choice)

		return nil
	},
}

func runVerificationChecks(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
	loadedAppConfig *config.Config,
	examplePath string,
) error {
	if loadedAppConfig.Run.Examples == nil {
		return nil
	}

	exampleSettings, ok := loadedAppConfig.Run.Examples[examplePath]
	if !ok || len(exampleSettings.Verify) == 0 {
		return nil
	}

	presenter.Header("--- 🔍 Verifying Prerequisites for '%s' ---", examplePath)

	allPassed := true

	for _, check := range exampleSettings.Verify {
		_, stderr, err := execClient.CaptureOutput(ctx, ".", check.Command, check.Args...)
		if err != nil {
			allPassed = false

			presenter.Error("  ❌ FAILED: Command '%s' failed.", check.Command)

			if stderr != "" {
				presenter.Detail("    Stderr: %s", stderr)
			}
		} else {
			presenter.Success("  ✅ PASSED")
		}
	}

	if !allPassed {
		return errors.New("verification failed")
	}

	return nil
}

func findRunnableExamples(rootDir string) ([]string, error) {
	var examples []string

	examplesDir := filepath.Join(rootDir, "examples")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read examples directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			examples = append(examples, filepath.ToSlash(filepath.Join("examples", entry.Name())))
		}
	}

	return examples, nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(runLongDescription, nil)
	if err != nil {
		panic(err)
	}

	RunCmd.Short = desc.Short
	RunCmd.Long = desc.Long
}
