// cmd/product/clean/clean.go
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

// CleanCmd represents the clean command
var CleanCmd = &cobra.Command{
	Use: "clean",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Header("--- Cleaning Local Project Files ---")

		dirsToRemove := []string{"./bin"}
		for _, dir := range dirsToRemove {
			presenter.Step("Removing directory: %s...", dir)
			if err := os.RemoveAll(dir); err != nil {
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
			if err := os.Remove(file); err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove file %s: %w", file, err)
				}
			}
		}

		presenter.Step("Cleaning Go build and test caches...")
		if err := globals.ExecClient.Execute(ctx, ".", "go", "clean", "-cache", "-testcache"); err != nil {
			return fmt.Errorf("failed to run 'go clean': %w", err)
		}

		presenter.Newline()
		presenter.Success("Project files cleaned successfully.")
		globals.AppLogger.InfoContext(ctx, "Project cleanup completed.")
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(cleanLongDescription, nil)
	if err != nil {
		panic(err)
	}
	CleanCmd.Short = desc.Short
	CleanCmd.Long = desc.Long
}
