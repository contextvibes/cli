// Package exportupstream provides the command to export upstream module source code.
package exportupstream

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed exportupstream.md.tpl
var exportUpstreamLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var outputFlag string

// ExportUpstreamCmd represents the export-upstream command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ExportUpstreamCmd = &cobra.Command{
	Use:   "export-upstream",
	Short: "Exports source code of upstream dependencies defined in config.",
	//nolint:lll // Long description.
	Long: `Scans the Go module cache for dependencies listed in .contextvibes.yaml (under project.upstreamModules) and exports their source code to a single file for AI context.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(os.Stdout, os.Stderr)
		ctx := context.Background()

		modules := globals.LoadedAppConfig.Project.UpstreamModules
		if len(modules) == 0 {
			presenter.Warning("No upstream modules configured in .contextvibes.yaml.")
			presenter.Advice("Add modules to 'project.upstreamModules' list in your config file.")

			return nil
		}

		presenter.Summary("Exporting Upstream Context")
		presenter.Info("Target modules: %v", modules)

		var outputBuffer bytes.Buffer
		fmt.Fprintf(&outputBuffer, "--- Upstream Context Export ---\n")
		fmt.Fprintf(&outputBuffer, "Generated at: %s\n\n", time.Now().Format(time.RFC3339))

		for _, mod := range modules {
			presenter.Step("Processing %s...", mod)

			// 1. Get Module Details (Dir and Version)
			modDir, modVer, err := resolveModule(ctx, mod)

			// If error OR empty directory, attempt to download
			if err != nil || modDir == "" {
				presenter.Info("  Module content not found locally. Attempting to download...")
				downloadErr := globals.ExecClient.Execute(ctx, ".", "go", "mod", "download", mod)
				if downloadErr != nil {
					presenter.Warning("  Failed to download %s: %v", mod, downloadErr)

					continue
				}

				// Retry resolution
				modDir, modVer, err = resolveModule(ctx, mod)
				if err != nil {
					presenter.Warning("  Skipping %s: %v", mod, err)

					continue
				}
			}

			if modDir == "" {
				presenter.Warning("  Skipping %s: Go reported an empty directory path even after download.", mod)

				continue
			}

			presenter.Detail("Location: %s", modDir)
			presenter.Detail("Version:  %s", modVer)

			// 2. Walk and Collect Files
			err = filepath.WalkDir(modDir, func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					// Skip vendor directories inside the module
					if entry.Name() == "vendor" {
						return fs.SkipDir
					}

					return nil
				}

				// Filter: Must be .go
				if !strings.HasSuffix(path, ".go") {
					return nil
				}

				// Filter: Skip tests, UNLESS it is an example test
				isTest := strings.HasSuffix(path, "_test.go")
				isExample := strings.HasPrefix(entry.Name(), "example") && isTest

				if isTest && !isExample {
					return nil
				}

				// Read content
				//nolint:gosec // Reading source files is intended.
				content, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("failed to read %s: %w", path, err)
				}

				// Calculate relative path for display
				relPath, _ := filepath.Rel(modDir, path)

				// Append to buffer using standard format
				fmt.Fprintf(&outputBuffer, "\n================================================\n")
				fmt.Fprintf(&outputBuffer, "MODULE: %s (%s)\n", mod, modVer)
				fmt.Fprintf(&outputBuffer, "FILE:   %s\n", relPath)
				fmt.Fprintf(&outputBuffer, "================================================\n")
				outputBuffer.Write(content)
				outputBuffer.WriteString("\n")

				return nil
			})

			if err != nil {
				presenter.Error("  Failed to walk module directory: %v", err)
			}
		}

		// Write to file
		err := tools.WriteBufferToFile(outputFlag, &outputBuffer)
		if err != nil {
			//nolint:wrapcheck // Wrapping is handled by caller.
			return fmt.Errorf("failed to write output file: %w", err)
		}

		presenter.Success("Context exported to: %s", outputFlag)

		return nil
	},
}

// resolveModule runs 'go list' to find the directory and version of a module.
func resolveModule(ctx context.Context, mod string) (string, string, error) {
	// Use -mod=readonly to force looking at the module cache/go.mod instead of vendor directory
	//nolint:lll // Command line arguments are long.
	out, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "go", "list", "-mod=readonly", "-m", "-f", "{{.Dir}}|{{.Version}}", mod)
	if err != nil {
		//nolint:wrapcheck // Wrapping is handled by caller.
		return "", "", err
	}

	parts := strings.Split(strings.TrimSpace(out), "|")
	//nolint:mnd // 2 parts expected: Dir and Version.
	if len(parts) != 2 {
		//nolint:err113 // Dynamic error is appropriate here.
		return "", "", fmt.Errorf("unexpected go list output format: %s", out)
	}

	return parts[0], parts[1], nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	// Create a default description if the template file is missing or empty during dev
	desc := cmddocs.CommandDesc{
		Short: "Exports source code of upstream dependencies.",
		Long:  "Scans and exports upstream Go module source code defined in configuration.",
	}

	parsed, err := cmddocs.ParseAndExecute(exportUpstreamLongDescription, nil)
	if err == nil {
		desc = parsed
	}

	ExportUpstreamCmd.Short = desc.Short
	ExportUpstreamCmd.Long = desc.Long

	ExportUpstreamCmd.Flags().StringVarP(&outputFlag, "output", "o", "upstream_context.txt", "Output file path")
}
