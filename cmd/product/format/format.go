// Package format provides the command to auto-format project source code.
package format

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed format.md.tpl
var formatLongDescription string

var errFormattingFailed = errors.New("one or more formatting tools failed")

// FormatCmd represents the format command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var FormatCmd = &cobra.Command{
	Use: "format [paths...]",
	Example: `  contextvibes product format                  # Format entire project
  contextvibes product format cmd/factory/scrub # Format specific package`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Applying code formatting and auto-fixes.")

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

		var formatErrors []error

		//nolint:exhaustive // We only handle supported project types, others fall to default.
		switch projType {
		case project.Go:
			presenter.Header("Go Formatting & Lint Fixes")

			// Construct arguments for golangci-lint
			lintArgs := []string{"run", "--fix"}
			if len(args) > 0 {
				lintArgs = append(lintArgs, args...)
			}

			err := runFormatCommand(
				ctx,
				presenter,
				globals.ExecClient,
				cwd,
				"golangci-lint",
				lintArgs,
			)
			if err != nil {
				presenter.Warning(
					"'golangci-lint --fix' completed but may have found unfixable issues.",
				)
			} else {
				presenter.Success("✓ golangci-lint completed.")
			}
			// Add other project types here later
		default:
			presenter.Info("No formatters configured for %s", projType)
		}

		presenter.Newline()
		if len(formatErrors) > 0 {
			return errFormattingFailed
		}

		presenter.Success("All formatting and auto-fixing tools completed.")

		return nil
	},
}

func runFormatCommand(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
	cwd, command string,
	args []string,
) error {
	presenter.Step("Running %s...", command)

	if !execClient.CommandExists(command) {
		presenter.Warning("'%s' command not found, skipping.", command)

		return nil
	}

	err := execClient.Execute(ctx, cwd, command, args...)
	if err != nil {
		return fmt.Errorf("failed to execute %s: %w", command, err)
	}

	return nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(formatLongDescription, nil)
	if err != nil {
		panic(err)
	}

	FormatCmd.Short = desc.Short
	FormatCmd.Long = desc.Long
}
