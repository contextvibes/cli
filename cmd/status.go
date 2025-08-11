// cmd/status.go

package cmd

import (
	"bufio" // For scanning output lines
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings" // For trimming space

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui" // Import presenter
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Shows a concise summary of the working tree status.",
	Long: `Displays a concise summary of the Git working tree status using 'git status --short'.
This includes staged changes, unstaged changes, and untracked files.`,
	Example: `  contextvibes status`,
	Args:    cobra.NoArgs,
	// Add Silence flags as we handle output/errors via Presenter
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Status: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "status"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "status"))

		// --- Get and Display Short Status ---
		presenter.Summary("Displaying Git repository status summary.") // User info

		logger.DebugContext(ctx, "Fetching short status", slog.String("source_command", "status"))
		stdout, stderr, err := client.GetStatusShort(ctx)

		// Log stderr from git command if any (usually empty for status --short unless error)
		if stderr != "" {
			logger.WarnContext(ctx, "stderr received from 'git status --short'",
				slog.String("source_command", "status"),
				slog.String("stderr", strings.TrimSpace(stderr)),
			)
		}

		// Handle execution errors
		if err != nil {
			presenter.Error("Failed to retrieve Git status: %v", err)
			// GetStatusShort already logged details
			return err
		}

		// --- Present the Status Output ---
		trimmedStdout := strings.TrimSpace(stdout)
		if trimmedStdout == "" {
			presenter.Info("Working tree is clean.") // Use Info for clean status
			logger.InfoContext(ctx, "Status check reported clean working tree", slog.String("source_command", "status"))
		} else {
			// Use the Info block to display the short status lines
			presenter.InfoPrefixOnly()                                                  // Print "INFO:" prefix
			_, _ = fmt.Fprintln(presenter.Out(), "  Current Changes (--short format):") // Add context header
			scanner := bufio.NewScanner(strings.NewReader(trimmedStdout))
			for scanner.Scan() {
				// Print each line indented under the INFO block
				_, _ = fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
			presenter.Newline() // Add newline after the block
			logger.InfoContext(ctx, "Status check reported changes", slog.String("source_command", "status"), slog.Int("line_count", strings.Count(trimmedStdout, "\n")+1))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
