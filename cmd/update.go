// cmd/update.go
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Finds all Go modules in the project and updates their dependencies.",
	Long: `Recursively searches for all 'go.mod' files in the current project
and runs 'go get -u ./...' and 'go mod tidy' in each module directory to update
all dependencies to their latest versions.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		logger := AppLogger
		ctx := cmd.Context()

		presenter.Header("--- Updating Go Module Dependencies ---")

		modules, err := findGoModules(".")
		if err != nil {
			presenter.Error("Error while searching for Go modules: %v", err)
			return err
		}

		if len(modules) == 0 {
			presenter.Info("No Go modules found. Nothing to do.")
			return nil
		}

		presenter.Info("Found %d Go module(s):", len(modules))
		for _, modDir := range modules {
			presenter.Detail("  - %s", modDir)
		}
		presenter.Newline()

		for _, modDir := range modules {
			presenter.Header("--- Processing module: %s ---", modDir)

			presenter.Step("Tidying go.mod and go.sum files...")
			if err := ExecClient.Execute(ctx, modDir, "go", "mod", "tidy"); err != nil {
				presenter.Error("Failed to run 'go mod tidy' in %s: %v", modDir, err)
				return fmt.Errorf("failed to run 'go mod tidy' in %s: %w", modDir, err)
			}

			presenter.Step("Updating dependencies to latest versions...")
			if err := ExecClient.Execute(ctx, modDir, "go", "get", "-u", "./..."); err != nil {
				presenter.Error("Failed to run 'go get -u' in %s: %v", modDir, err)
				return fmt.Errorf("failed to run 'go get -u' in %s: %w", modDir, err)
			}

			presenter.Step("Tidying again after updates...")
			if err := ExecClient.Execute(ctx, modDir, "go", "mod", "tidy"); err != nil {
				presenter.Error("Failed to run final 'go mod tidy' in %s: %v", modDir, err)
				return fmt.Errorf("failed to run final 'go mod tidy' in %s: %w", modDir, err)
			}
		}

		presenter.Newline()
		presenter.Success("All Go modules updated successfully.")
		logger.InfoContext(ctx, "Go module update completed.")
		return nil
	},
}

// findGoModules recursively finds directories containing a go.mod file.
func findGoModules(rootDir string) ([]string, error) {
	var modules []string
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "vendor" || d.Name() == ".git" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "go.mod" {
			modules = append(modules, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directories to find modules: %w", err)
	}
	return modules, nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
