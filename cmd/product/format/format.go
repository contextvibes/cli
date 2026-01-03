// Package format provides the command to auto-format project source code.
package format

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed format.md.tpl
var formatLongDescription string

var (
	errFormattingFailed = errors.New("one or more formatting tools failed")
	// nolintRegex matches "//nolint" or "// nolint" and anything following it on the line.
	nolintRegex = regexp.MustCompile(`//\s*nolint.*`)
)

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	removeDirectives bool
	strictMode       bool
)

// FormatCmd represents the format command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var FormatCmd = &cobra.Command{
	Use: "format [paths...]",
	Example: `  contextvibes product format                  # Format entire project
  contextvibes product format --strict         # Format using strict rules
  contextvibes product format --remove-directives # Remove //nolint comments
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

		// Initialize Git Client early for safety checks and status reporting
		//nolint:exhaustruct // Partial config is sufficient.
		gitCfg := git.GitClientConfig{
			Logger:   globals.AppLogger,
			Executor: globals.ExecClient.UnderlyingExecutor(),
		}
		gitClient, err := git.NewClient(ctx, cwd, gitCfg)
		if err != nil {
			presenter.Warning("Could not initialize Git client. Safety checks and smart reporting will be disabled.")
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))

		// --- Report Buffer Initialization ---
		var reportBuffer bytes.Buffer
		fmt.Fprintf(&reportBuffer, "# Format Report (%s)\n\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(&reportBuffer, "> **AI Instruction:** This report details the results of the `format` command. Review the 'Execution Log' for errors. If files are listed under 'Updated File Content', treat that content as the new Ground Truth, superseding previous context. If no files are listed there, no changes were made.\n\n")
		fmt.Fprintf(&reportBuffer, "## Execution Log\n\n")

		var formatErrors []error

		//nolint:exhaustive // We only handle supported project types, others fall to default.
		switch projType {
		case project.Go:
			// Optional: Remove Linter Directives
			if removeDirectives {
				// Pass gitClient to reuse it
				modified, err := runStripDirectives(ctx, presenter, gitClient, cwd, args)
				if err != nil {
					return err
				}
				if len(modified) > 0 {
					fmt.Fprintf(&reportBuffer, "### Strip Directives\n")
					for _, f := range modified {
						fmt.Fprintf(&reportBuffer, "- Cleaned: %s\n", f)
					}
					fmt.Fprintf(&reportBuffer, "\n")
				}
			}

			presenter.Header("Go Formatting & Lint Fixes")

			// 1. Run go mod tidy
			if globals.ExecClient.CommandExists("go") {
				out, err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "go", []string{"mod", "tidy"})
				logOutput(&reportBuffer, "go mod tidy", out, err)
				if err != nil {
					formatErrors = append(formatErrors, err)
				} else {
					presenter.Success("✓ go mod tidy applied.")
				}
			}

			// 2. Run goimports (with -l to list files)
			if globals.ExecClient.CommandExists("goimports") {
				goimportsArgs := []string{"-l", "-w"} // -l lists files
				if len(args) > 0 {
					goimportsArgs = append(goimportsArgs, args...)
				} else {
					goimportsArgs = append(goimportsArgs, ".")
				}

				out, err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "goimports", goimportsArgs)
				logOutput(&reportBuffer, "goimports", out, err)

				if err != nil {
					formatErrors = append(formatErrors, err)
				} else {
					if len(out) > 0 {
						lines := strings.SplitSeq(strings.TrimSpace(out), "\n")
						for line := range lines {
							presenter.Detail("Updated: %s", line)
						}
					}
					presenter.Success("✓ goimports applied.")
				}
			} else {
				presenter.Warning("goimports not found. Install 'gotools' for better import management.")
			}

			// 3. Run gofmt (with -l to list files)
			gofmtArgs := []string{"-l", "-s", "-w"} // -l lists files
			if len(args) > 0 {
				gofmtArgs = append(gofmtArgs, args...)
			} else {
				gofmtArgs = append(gofmtArgs, ".")
			}

			out, err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "gofmt", gofmtArgs)
			logOutput(&reportBuffer, "gofmt", out, err)

			if err != nil {
				formatErrors = append(formatErrors, err)
			} else {
				if len(out) > 0 {
					lines := strings.SplitSeq(strings.TrimSpace(out), "\n")
					for line := range lines {
						presenter.Detail("Updated: %s", line)
					}
				}
				presenter.Success("✓ gofmt -s applied.")
			}

			// 4. Run golangci-lint --fix
			lintArgs := []string{"run", "--fix"}

			if strictMode {
				presenter.Info("Using strict configuration for linter fixes.")
				configBytes, err := config.GetLanguageAsset("go", config.AssetLintStrict)
				if err != nil {
					return fmt.Errorf("failed to load strict config asset: %w", err)
				}

				tmpFile, err := os.CreateTemp(".", ".golangci-strict-*.yml")
				if err != nil {
					return fmt.Errorf("failed to create temp strict config: %w", err)
				}
				defer os.Remove(tmpFile.Name())

				if _, err := tmpFile.Write(configBytes); err != nil {
					return fmt.Errorf("failed to write strict config: %w", err)
				}
				tmpFile.Close()

				lintArgs = append(lintArgs, "-c", tmpFile.Name())
			}

			if len(args) > 0 {
				lintArgs = append(lintArgs, args...)
			}

			lintOut, lintErr := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "golangci-lint", lintArgs)
			if strictMode {
				lintOut = strings.ReplaceAll(lintOut, ".golangci-strict-*.yml", ".golangci-strict.yml")
			}
			logOutput(&reportBuffer, "golangci-lint", lintOut, lintErr)

			if lintErr != nil {
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

		// --- Context Refresh: Output Content of MODIFIED Files Only ---
		if gitClient != nil {
			// Check git status to see what actually changed
			statusOut, _, err := gitClient.GetStatusShort(ctx)
			if err == nil && len(statusOut) > 0 {
				dirtyFiles := parseGitStatus(statusOut)

				// If specific args were provided, filter dirty files by those args
				// If no args (project wide), include all dirty files
				filesToDump := filterFiles(dirtyFiles, args)

				if len(filesToDump) > 0 {
					fmt.Fprintf(&reportBuffer, "\n## Updated File Content (Ground Truth)\n")
					for _, path := range filesToDump {
						// Double check it's a file and exists
						info, err := os.Stat(path)
						if err == nil && !info.IsDir() {
							content, err := os.ReadFile(path)
							if err == nil {
								tools.AppendFileMarkerHeader(&reportBuffer, path)
								reportBuffer.Write(content)
								tools.AppendFileMarkerFooter(&reportBuffer, path)
							}
						}
					}
				}
			}
		}

		// Write the full report to _contextvibes.md
		if err := tools.WriteBufferToFile("_contextvibes.md", &reportBuffer); err != nil {
			presenter.Error("Failed to write report to _contextvibes.md: %v", err)
		} else {
			presenter.Success("Full context written to %s", presenter.Highlight("_contextvibes.md"))
		}

		return nil
	},
}

func logOutput(buf *bytes.Buffer, tool string, output string, err error) {
	fmt.Fprintf(buf, "### %s\n", tool)

	if err != nil {
		fmt.Fprintf(buf, "**Status:** Failed\n")
		fmt.Fprintf(buf, "**Error:** %v\n", err)
	} else {
		fmt.Fprintf(buf, "**Status:** Success\n")
	}

	if len(output) > 0 {
		fmt.Fprintf(buf, "\n```text\n%s\n```\n", strings.TrimSpace(output))
	}
	// Removed redundant "(No output)" block
	fmt.Fprintf(buf, "\n")
}

// parseGitStatus parses "git status --short" output and returns a list of file paths.
// Example line: " M cmd/main.go" or "?? newfile.go".
func parseGitStatus(output string) []string {
	var files []string

	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Status is usually 2 chars, space, then path.
		// But TrimSpace removes leading spaces.
		// Simple heuristic: split by space, take the last part.
		// This handles standard cases. For quoted paths with spaces, this is a limitation we accept for now.
		parts := strings.Fields(trimmed)
		if len(parts) >= 2 {
			files = append(files, parts[len(parts)-1])
		}
	}

	return files
}

// filterFiles returns files from 'candidates' that match 'targets'.
// If targets is empty, returns all candidates.
func filterFiles(candidates []string, targets []string) []string {
	if len(targets) == 0 {
		return candidates
	}

	var result []string

	for _, c := range candidates {
		for _, t := range targets {
			// Simple check: is the candidate the target, or inside the target directory?
			// Normalize paths for comparison
			cleanC := filepath.Clean(c)
			cleanT := filepath.Clean(t)

			if cleanC == cleanT || strings.HasPrefix(cleanC, cleanT+string(os.PathSeparator)) {
				result = append(result, c)

				break
			}
		}
	}

	return result
}

func runStripDirectives(ctx context.Context, presenter *ui.Presenter, client *git.GitClient, cwd string, args []string) ([]string, error) {
	presenter.Header("Removing Linter Directives")

	// 1. Safety Check: Ensure Git is clean
	if client != nil {
		isClean, err := client.IsWorkingDirClean(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check git status: %w", err)
		}

		if !isClean {
			presenter.Warning("Your working directory is dirty (has uncommitted changes).")

			if globals.AssumeYes {
				presenter.Error("Cannot proceed in non-interactive mode with dirty directory.")
				//nolint:err113 // Dynamic error is appropriate here.
				return nil, errors.New("working directory not clean")
			}

			confirmCommit, err := presenter.PromptForConfirmation("Stage and commit all changes now to proceed?")
			if err != nil {
				return nil, fmt.Errorf("prompt failed: %w", err)
			}

			if !confirmCommit {
				presenter.Info("Aborted by user.")
				//nolint:err113 // Dynamic error is appropriate here.
				return nil, errors.New("user aborted on dirty directory")
			}

			presenter.Step("Staging and committing changes...")

			if err := client.AddAll(ctx); err != nil {
				return nil, fmt.Errorf("failed to stage changes: %w", err)
			}

			if err := client.Commit(ctx, "chore: Save state before removing directives"); err != nil {
				return nil, fmt.Errorf("failed to commit changes: %w", err)
			}

			presenter.Success("✓ State saved.")
		}
	}

	presenter.Warning("This will remove all '//nolint' directives from Go files.")

	if !globals.AssumeYes {
		confirmed, err := presenter.PromptForConfirmation("Are you sure you want to proceed?")
		if err != nil {
			return nil, fmt.Errorf("confirmation failed: %w", err)
		}

		if !confirmed {
			presenter.Info("Skipping directive removal.")

			return nil, nil
		}
	}

	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var modifiedFiles []string

	for _, path := range paths {
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				if d.Name() == "vendor" || d.Name() == ".git" {
					return fs.SkipDir
				}

				return nil
			}

			if strings.HasSuffix(path, ".go") {
				modified, err := stripFile(path)
				if err != nil {
					presenter.Error("Failed to process %s: %v", path, err)

					return nil // Continue processing other files
				}

				if modified {
					modifiedFiles = append(modifiedFiles, path)
					presenter.Detail("Cleaned %s", path)
				}
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking path %s: %w", path, err)
		}
	}

	presenter.Success("Removed directives from %d files.", len(modifiedFiles))

	return modifiedFiles, nil
}

func stripFile(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read failed: %w", err)
	}

	original := string(content)
	modified := nolintRegex.ReplaceAllString(original, "")

	if original == modified {
		return false, nil
	}

	//nolint:mnd // 0600 is standard file permission.
	if err := os.WriteFile(path, []byte(modified), 0o600); err != nil {
		return false, fmt.Errorf("write failed: %w", err)
	}

	return true, nil
}

func runFormatCommand(
	ctx context.Context,
	presenter *ui.Presenter,
	execClient *exec.ExecutorClient,
	cwd, command string,
	args []string,
) (string, error) {
	presenter.Step("Running %s %s...", command, args[0]) // Simple log

	if !execClient.CommandExists(command) {
		presenter.Warning("'%s' command not found, skipping.", command)

		return "", nil
	}

	// Use CaptureOutput instead of Execute to get the logs
	stdout, stderr, err := execClient.CaptureOutput(ctx, cwd, command, args...)
	output := stdout + stderr

	if err != nil {
		return output, fmt.Errorf("failed to execute %s: %w", command, err)
	}

	return output, nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(formatLongDescription, nil)
	if err != nil {
		panic(err)
	}

	FormatCmd.Short = desc.Short
	FormatCmd.Long = desc.Long

	FormatCmd.Flags().BoolVar(&removeDirectives, "remove-directives", false, "Remove all //nolint directives from Go files")
	FormatCmd.Flags().BoolVar(&strictMode, "strict", false, "Use strict configuration for linter fixes")
}
