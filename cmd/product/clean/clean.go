// Package clean provides the command to clean project artifacts.
package clean

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed clean.md.tpl
var cleanLongDescription string

// CleanCmd represents the clean command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes temporary files and build artifacts.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Header("--- Cleaning Local Project Files ---")

		dirsToRemove := []string{"./bin"}
		for _, dir := range dirsToRemove {
			presenter.Step("Removing directory: %s...", dir)
			err := os.RemoveAll(dir)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove directory %s: %w", dir, err)
				}
			}
		}

		filesToRemove := []string{
			"coverage.out",
			"context_commit.md",
			"context_pr.md",
			"context_export_project.md",
		}
		for _, file := range filesToRemove {
			presenter.Step("Removing file: %s...", file)
			err := os.Remove(file)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove file %s: %w", file, err)
				}
			}
		}

		presenter.Step("Cleaning Go build and test caches...")
		err := globals.ExecClient.Execute(ctx, ".", "go", "clean", "-cache", "-testcache")
		if err != nil {
			return fmt.Errorf("failed to run 'go clean': %w", err)
		}

		presenter.Newline()
		presenter.Success("Project files cleaned successfully.")
		globals.AppLogger.InfoContext(ctx, "Project cleanup completed.")

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(cleanLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CleanCmd.Short = desc.Short
	CleanCmd.Long = desc.Long
}
