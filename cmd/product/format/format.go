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
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))

		var formatErrors []error

		//nolint:exhaustive // We only handle supported project types, others fall to default.
		switch projType {
		case project.Go:
			presenter.Header("Go Formatting & Lint Fixes")

			// 1. Run goimports (if available) - Best for imports
			if globals.ExecClient.CommandExists("goimports") {
				goimportsArgs := []string{"-w"}
				if len(args) > 0 {
					goimportsArgs = append(goimportsArgs, args...)
				} else {
					goimportsArgs = append(goimportsArgs, ".")
				}

				// Uber Style: Inline error check reduces scope
				if err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "goimports", goimportsArgs); err != nil {
					formatErrors = append(formatErrors, err)
				} else {
					presenter.Success("✓ goimports applied.")
				}
			} else {
				presenter.Warning("goimports not found. Install 'gotools' for better import management.")
			}

			// 2. Run gofmt -s (Standard simplification)
			gofmtArgs := []string{"-s", "-w"}
			if len(args) > 0 {
				gofmtArgs = append(gofmtArgs, args...)
			} else {
				gofmtArgs = append(gofmtArgs, ".")
			}

			// Uber Style: 'err' is a new variable in this new 'if' scope
			if err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "gofmt", gofmtArgs); err != nil {
				formatErrors = append(formatErrors, err)
			} else {
				presenter.Success("✓ gofmt -s applied.")
			}

			// 3. Run golangci-lint --fix (Deep fixes)
			lintArgs := []string{"run", "--fix"}
			if len(args) > 0 {
				lintArgs = append(lintArgs, args...)
			}

			// Uber Style: Again, 'err' is scoped to this block
			if err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "golangci-lint", lintArgs); err != nil {
				presenter.Warning("'golangci-lint --fix' completed but may have found unfixable issues.")
			} else {
				presenter.Success("✓ golangci-lint --fix applied.")
			}

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

	if err := execClient.Execute(ctx, cwd, command, args...); err != nil {
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
