// cmd/codemod.go
package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/contextvibes/cli/internal/codemod" // Using the new types package
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

const defaultCodemodFilename = "codemod.json" // Default script filename

var codemodScriptPath string

var codemodCmd = &cobra.Command{
	Use:   "codemod [--script <file.json>]",
	Short: "Applies programmatic code modifications or deletions from a JSON script.",
	Long: `Reads a JSON script file describing a series of operations to be applied to
specified files in the codebase. This enables automated or AI-assisted refactoring and cleanup.

If --script is not provided, it looks for '` + defaultCodemodFilename + `' in the current directory.

The JSON script should be an array of objects, where each object defines:
  - "file_path": The path to the file to be modified or deleted.
  - "operations": An array of operation objects for that file.

Currently supported operation types:
  - "regex_replace": Performs find/replace on file content.
    Required fields: "type": "regex_replace", "find_regex": "...", "replace_with": "..."
  - "delete_file": Deletes the specified file_path.
    Required fields: "type": "delete_file"

Requires confirmation before writing/deleting, unless --yes is specified.`,
	Example: `  contextvibes codemod # Looks for codemod.json
  contextvibes codemod --script ./my_refactor_script.json
  contextvibes codemod -s ./cleanup.json -y`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)

		scriptToLoad := codemodScriptPath

		if scriptToLoad == "" {
			presenter.Info("No --script provided, attempting to load default: %s", defaultCodemodFilename)
			if _, err := os.Stat(defaultCodemodFilename); os.IsNotExist(err) {
				presenter.Error("Default codemod script '%s' not found and no --script flag provided.", defaultCodemodFilename)
				presenter.Advice("Create '%s' in the current directory or use the --script flag to specify a file.", defaultCodemodFilename)
				return fmt.Errorf("no codemod script specified or found")
			}
			scriptToLoad = defaultCodemodFilename
		}

		presenter.Summary("Applying codemod script: %s", scriptToLoad)

		scriptData, err := os.ReadFile(scriptToLoad)
		if err != nil {
			presenter.Error("Failed to read codemod script file '%s': %v", scriptToLoad, err)
			logger.Error("codemod: failed to read script file", slog.String("path", scriptToLoad), slog.Any("error", err))
			return err
		}

		var script codemod.ChangeScript
		if err := json.Unmarshal(scriptData, &script); err != nil {
			presenter.Error("Failed to parse codemod script JSON from '%s': %v", scriptToLoad, err)
			logger.Error("codemod: failed to parse script json", slog.String("path", scriptToLoad), slog.Any("error", err))
			return err
		}

		if len(script) == 0 {
			presenter.Info("Codemod script is empty. No changes to apply.")
			return nil
		}

		totalFilesModified := 0
		totalFilesDeleted := 0
		totalOperationsAttempted := 0
		totalOperationsSucceeded := 0

		for _, fileChangeSet := range script {
			presenter.Header("Processing target: %s", fileChangeSet.FilePath)

			onlyDelete := len(fileChangeSet.Operations) == 1 && fileChangeSet.Operations[0].Type == "delete_file"
			fileExists := false
			fileInfo, statErr := os.Stat(fileChangeSet.FilePath) // Stat once
			if statErr == nil {
				fileExists = true
			} else if !os.IsNotExist(statErr) {
				presenter.Error("Error checking file %s: %v. Skipping.", fileChangeSet.FilePath, statErr)
				logger.Error("codemod: failed to stat file", slog.String("path", fileChangeSet.FilePath), slog.Any("error", statErr))
				continue
			}

			if !fileExists && !onlyDelete {
				presenter.Error("File not found: %s. Skipping operations (except delete_file).", fileChangeSet.FilePath)
				logger.Error("codemod: file not found, skipping changeset", slog.String("path", fileChangeSet.FilePath))
				continue
			}

			var currentFileContent string
			var contentBeforeOpsForThisFile string
			if fileExists && !onlyDelete {
				fileContentBytes, readErr := os.ReadFile(fileChangeSet.FilePath)
				if readErr != nil {
					presenter.Error("Failed to read file %s: %v. Skipping.", fileChangeSet.FilePath, readErr)
					logger.Error("codemod: failed to read file for modification", slog.String("path", fileChangeSet.FilePath), slog.Any("error", readErr))
					continue
				}
				currentFileContent = string(fileContentBytes)
				contentBeforeOpsForThisFile = currentFileContent
			} else {
				contentBeforeOpsForThisFile = ""
				currentFileContent = ""
			}

			fileWasDeleted := false

			for opIndex, op := range fileChangeSet.Operations {
				totalOperationsAttempted++
				presenter.Step("Attempting operation %d: type='%s', desc='%s'", opIndex+1, op.Type, op.Description)

				opSucceeded := false
				contentBeforeThisOp := currentFileContent

				switch op.Type {
				case "regex_replace":
					if fileWasDeleted {
						presenter.Warning("Skipping regex_replace on '%s' as file was already deleted by a previous operation.", fileChangeSet.FilePath)
						continue
					}
					if !fileExists {
						presenter.Error("Cannot apply regex_replace: file '%s' does not exist.", fileChangeSet.FilePath)
						continue
					}
					if op.FindRegex == "" {
						presenter.Warning("Skipping regex_replace for '%s': find_regex is empty.", fileChangeSet.FilePath)
						logger.Warn("codemod: regex_replace skipped, empty find_regex", slog.String("file", fileChangeSet.FilePath))
						continue
					}
					re, compileErr := regexp.Compile(op.FindRegex)
					if compileErr != nil {
						presenter.Error("Invalid regex '%s' for file '%s': %v. Skipping operation.", op.FindRegex, fileChangeSet.FilePath, compileErr)
						logger.Error("codemod: invalid regex in script", slog.String("file", fileChangeSet.FilePath), slog.String("regex", op.FindRegex), slog.Any("error", compileErr))
						continue
					}

					currentFileContent = re.ReplaceAllString(currentFileContent, op.ReplaceWith)
					if currentFileContent != contentBeforeThisOp {
						opSucceeded = true
						presenter.Info("  Applied regex_replace to '%s'.", fileChangeSet.FilePath)
						logger.Info("codemod: applied regex_replace", slog.String("file", fileChangeSet.FilePath), slog.String("find", op.FindRegex), slog.String("replace", op.ReplaceWith))
					} else {
						opSucceeded = true
						presenter.Info("  Regex_replace on '%s' resulted in no changes.", fileChangeSet.FilePath)
						logger.Info("codemod: regex_replace no change", slog.String("file", fileChangeSet.FilePath), slog.String("find", op.FindRegex))
					}

				case "delete_file":
					if !fileExists {
						presenter.Info("File '%s' already does not exist. 'delete_file' operation considered successful.", fileChangeSet.FilePath)
						opSucceeded = true
						fileWasDeleted = true
						break // Break from inner loop
					}
					if fileWasDeleted {
						presenter.Info("File '%s' already marked for deletion by previous operation.", fileChangeSet.FilePath)
						opSucceeded = true
						break // Break from inner loop
					}

					presenter.Info("Operation requests deletion of file: %s", fileChangeSet.FilePath)
					deleteConfirmed := false
					if assumeYes {
						presenter.Info("Deleting file '%s' (confirmation bypassed via --yes).", fileChangeSet.FilePath)
						deleteConfirmed = true
					} else {
						var promptErr error
						deleteConfirmed, promptErr = presenter.PromptForConfirmation(fmt.Sprintf("Permanently delete file '%s'?", fileChangeSet.FilePath))
						if promptErr != nil {
							presenter.Error("Error during delete confirmation for '%s': %v. Skipping deletion.", fileChangeSet.FilePath, promptErr)
							logger.Error("codemod: delete confirmation error", slog.String("file", fileChangeSet.FilePath), slog.Any("error", promptErr))
							continue // Skip op
						}
					}

					if deleteConfirmed {
						err := os.Remove(fileChangeSet.FilePath)
						if err != nil {
							if os.IsNotExist(err) {
								presenter.Warning("File '%s' did not exist when deletion was attempted.", fileChangeSet.FilePath)
								opSucceeded = true
								fileWasDeleted = true
							} else {
								presenter.Error("Failed to delete file '%s': %v", fileChangeSet.FilePath, err)
								logger.Error("codemod: failed to delete file", slog.String("path", fileChangeSet.FilePath), slog.Any("error", err))
							}
						} else {
							presenter.Success("Successfully deleted file: %s", fileChangeSet.FilePath)
							logger.Info("codemod: deleted file", slog.String("path", fileChangeSet.FilePath))
							opSucceeded = true
							fileWasDeleted = true
							totalFilesDeleted++
						}
					} else {
						presenter.Info("Skipped deletion of file '%s' by user.", fileChangeSet.FilePath)
						logger.Info("codemod: delete skipped by user", slog.String("file", fileChangeSet.FilePath))
					}
					if fileWasDeleted {
						break // exit the inner operations loop for this file
					}


				default:
					presenter.Warning("Unsupported operation type: '%s' for file '%s'. Skipping.", op.Type, fileChangeSet.FilePath)
					logger.Warn("codemod: unsupported operation type", slog.String("type", op.Type), slog.String("file", fileChangeSet.FilePath))
				} // end switch op.Type

				if opSucceeded {
					totalOperationsSucceeded++
				}
			} // end operations loop

			if !fileWasDeleted && currentFileContent != contentBeforeOpsForThisFile {
				presenter.Info("File '%s' has pending modifications.", fileChangeSet.FilePath)

				confirmed := false
				if assumeYes {
					presenter.Info("Writing changes to '%s' (confirmation bypassed via --yes).", fileChangeSet.FilePath)
					confirmed = true
				} else {
					var promptErr error
					confirmed, promptErr = presenter.PromptForConfirmation(fmt.Sprintf("Write modified content to '%s'?", fileChangeSet.FilePath))
					if promptErr != nil {
						presenter.Error("Error during confirmation for '%s': %v. Skipping write.", fileChangeSet.FilePath, promptErr)
						logger.Error("codemod: confirmation error", slog.String("file", fileChangeSet.FilePath), slog.Any("error", promptErr))
						continue
					}
				}

				if confirmed {
					var perm os.FileMode = 0644
					if fileInfo != nil { // Use the FileInfo from the initial Stat
						perm = fileInfo.Mode().Perm()
					} else {
						// Should only happen if file didn't exist initially, but we are somehow trying to write... defensive log.
						logger.Warn("codemod: could not get original file permissions, using default 0644", slog.String("file", fileChangeSet.FilePath))
					}

					err := os.WriteFile(fileChangeSet.FilePath, []byte(currentFileContent), perm)
					if err != nil {
						presenter.Error("Failed to write changes to %s: %v", fileChangeSet.FilePath, err)
						logger.Error("codemod: failed to write file", slog.String("path", fileChangeSet.FilePath), slog.Any("error", err))
					} else {
						presenter.Success("Successfully updated %s", fileChangeSet.FilePath)
						totalFilesModified++
					}
				} else {
					presenter.Info("Skipped writing changes to %s due to user cancellation.", fileChangeSet.FilePath)
					logger.Info("codemod: write skipped by user", slog.String("file", fileChangeSet.FilePath))
				}
			} else if !fileWasDeleted {
				presenter.Info("No effective changes made to %s after all operations.", fileChangeSet.FilePath)
			}
			presenter.Newline()
		} // end fileChangeSet loop

		presenter.Separator()
		presenter.Summary("Codemod script execution finished.")
		presenter.Detail("Files Modified: %d", totalFilesModified)
		presenter.Detail("Files Deleted: %d", totalFilesDeleted)
		presenter.Detail("Operations Attempted: %d", totalOperationsAttempted)
		presenter.Detail("Operations Succeeded (may include no-ops/already deleted): %d", totalOperationsSucceeded)
		return nil
	},
}

func init() {
	codemodCmd.Flags().StringVarP(&codemodScriptPath, "script", "s", "", "Path to the JSON codemod script file (default: "+defaultCodemodFilename+")")
	// Flag is optional now
	rootCmd.AddCommand(codemodCmd)
}
