// cmd/clean.go
package cmd

import (
	"fmt"
	"os"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes temporary files, build artifacts, and caches.",
	Long: `Cleans the project directory by removing common temporary files,
build artifacts, and local caches.

This includes the './bin/' directory, Go test and build caches, and generated
context files like 'coverage.out' and 'context_*.md'.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := cmd.Context()
		logger := AppLogger

		presenter.Header("--- Cleaning Local Project Files ---")

		dirsToRemove := []string{"./bin"}
		for _, dir := range dirsToRemove {
			presenter.Step("Removing directory: %s...", dir)
			if err := os.RemoveAll(dir); err != nil {
				if !os.IsNotExist(err) {
					presenter.Error("Failed to remove directory %s: %v", dir, err)
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
					presenter.Error("Failed to remove file %s: %v", file, err)
					return fmt.Errorf("failed to remove file %s: %w", file, err)
				}
			}
		}

		presenter.Step("Cleaning Go build and test caches...")
		if err := ExecClient.Execute(ctx, ".", "go", "clean", "-cache", "-testcache"); err != nil {
			presenter.Error("Failed to run 'go clean': %v", err)
			return fmt.Errorf("failed to run 'go clean': %w", err)
		}

		presenter.Newline()
		presenter.Success("Project files cleaned successfully.")
		logger.InfoContext(ctx, "Project cleanup completed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
