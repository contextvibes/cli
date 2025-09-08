// cmd/product/format/format.go
package format

import (
	"context"
	_ "embed"
	"errors"
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

// FormatCmd represents the format command
var FormatCmd = &cobra.Command{
	Use:     "format",
	Example: `  contextvibes product format  # Apply formatting and fixes to the project`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Applying code formatting and auto-fixes.")

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

		var formatErrors []error

		switch projType {
		case project.Go:
			presenter.Header("Go Formatting & Lint Fixes")
			err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "golangci-lint", []string{"run", "--fix"})
			if err != nil {
				presenter.Warning("'golangci-lint --fix' completed but may have found unfixable issues.")
			} else {
				presenter.Success("✓ golangci-lint completed.")
			}
			// Add other project types here later
		}

		presenter.Newline()
		if len(formatErrors) > 0 {
			return errors.New("one or more formatting tools failed")
		}

		presenter.Success("All formatting and auto-fixing tools completed.")
		return nil
	},
}

func runFormatCommand(ctx context.Context, presenter *ui.Presenter, execClient *exec.ExecutorClient, cwd, command string, args []string) error {
	presenter.Step("Running %s...", command)
	if !execClient.CommandExists(command) {
		presenter.Warning("'%s' command not found, skipping.", command)
		return nil
	}
	return execClient.Execute(ctx, cwd, command, args...)
}

func init() {
	desc, err := cmddocs.ParseAndExecute(formatLongDescription, nil)
	if err != nil {
		panic(err)
	}
	FormatCmd.Short = desc.Short
	FormatCmd.Long = desc.Long
}
