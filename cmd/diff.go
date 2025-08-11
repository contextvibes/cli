// cmd/diff.go

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/git"   // Use GitClient
	"github.com/contextvibes/cli/internal/tools" // For Markdown/File IO
	"github.com/contextvibes/cli/internal/ui"    // Use Presenter for UI
	"github.com/spf13/cobra"
)

const fixedDiffOutputFile = "contextvibes.md" // Keep specific to diff command

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: fmt.Sprintf("Shows pending Git changes, OVERWRITING %s.", fixedDiffOutputFile),
	Long: fmt.Sprintf(`Generates a Markdown summary of pending changes (staged, unstaged, untracked)
in the Git repository and OVERWRITES the context file: %s.

This is useful for providing diff context to AI assistants or for quick status checks.
Run 'contextvibes describe' again if you need the full project context instead.`, fixedDiffOutputFile),
	Example: `  contextvibes diff  # OVERWRITES contextvibes.md with diff summary`,
	Args:    cobra.NoArgs,
	// Add Silence flags as we handle output/errors via Presenter
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return errors.New("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Generating Git diff summary for %s.", fixedDiffOutputFile)

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Diff: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed getwd: %v", err)

			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "diff"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)

			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "diff"))

		// --- Generate Diff Content ---
		var outputBuffer bytes.Buffer // Buffer to build markdown content
		var hasChanges bool

		presenter.Step("Checking for staged changes...") // Use Step
		stagedOut, _, stagedErr := client.GetDiffCached(ctx)
		if stagedErr != nil {
			presenter.Error("Failed to get staged changes: %v", stagedErr)

			return stagedErr // Client logs details
		}
		stagedOut = strings.TrimSpace(stagedOut)
		if stagedOut != "" {
			hasChanges = true
			logger.DebugContext(ctx, "Adding staged changes to buffer", slog.String("source_command", "diff"))
			tools.AppendSectionHeader(&outputBuffer, "Staged Changes (Index / `git diff --cached`)")
			tools.AppendFencedCodeBlock(&outputBuffer, stagedOut, "diff")
		}

		presenter.Step("Checking for unstaged changes...") // Use Step
		unstagedOut, _, unstagedErr := client.GetDiffUnstaged(ctx)
		if unstagedErr != nil {
			presenter.Error("Failed to get unstaged changes: %v", unstagedErr)

			return unstagedErr // Client logs details
		}
		unstagedOut = strings.TrimSpace(unstagedOut)
		if unstagedOut != "" {
			hasChanges = true
			logger.DebugContext(ctx, "Adding unstaged changes to buffer", slog.String("source_command", "diff"))
			tools.AppendSectionHeader(&outputBuffer, "Unstaged Changes (Working Directory / `git diff HEAD`)")
			tools.AppendFencedCodeBlock(&outputBuffer, unstagedOut, "diff")
		}

		presenter.Step("Checking for untracked files...") // Use Step
		untrackedOut, _, untrackedErr := client.ListUntrackedFiles(ctx)
		if untrackedErr != nil {
			presenter.Error("Failed to list untracked files: %v", untrackedErr)

			return untrackedErr // Client logs details
		}
		untrackedOut = strings.TrimSpace(untrackedOut)
		if untrackedOut != "" {
			hasChanges = true
			logger.DebugContext(ctx, "Adding untracked files to buffer", slog.String("source_command", "diff"))
			tools.AppendSectionHeader(&outputBuffer, "Untracked Files (`git ls-files --others --exclude-standard`)")
			tools.AppendFencedCodeBlock(&outputBuffer, untrackedOut, "")
		}

		// --- Write Output File or Report No Changes ---
		presenter.Newline()
		if !hasChanges {
			presenter.Info("No pending changes found.")
			presenter.Advice("The context file '%s' remains unchanged.", fixedDiffOutputFile)
			logger.InfoContext(ctx, "No pending git changes detected.", slog.String("source_command", "diff"))
		} else {
			presenter.Step("Writing diff summary, overwriting %s...", fixedDiffOutputFile) // Use Step

			// tools.WriteBufferToFile currently prints its own messages.
			// If we want full control via Presenter, WriteBufferToFile would need
			// modification or replacement. Let's keep it for now.
			errWrite := tools.WriteBufferToFile(fixedDiffOutputFile, &outputBuffer)
			if errWrite != nil {
				presenter.Error("Failed to write output file '%s': %v", fixedDiffOutputFile, errWrite)
				logger.ErrorContext(ctx, "Failed to write diff output file" /*...*/)

				return errWrite
			}
			// Success message is printed by WriteBufferToFile currently.
			// If WriteBufferToFile is made silent:
			// presenter.Success("Successfully wrote diff summary to %s.", fixedDiffOutputFile)
			logger.InfoContext(ctx, "Successfully wrote git diff summary to file." /*...*/)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
