// Package diff provides the command to show git diffs.
package diff

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed diff.md.tpl
var diffLongDescription string

// DiffCmd represents the diff command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var DiffCmd = &cobra.Command{
	Use:     "diff",
	Example: `  contextvibes factory diff`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Generating Git diff summary for %s.", config.DefaultDescribeOutputFile)

		workDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		//nolint:exhaustruct // Partial config is sufficient.
		gitCfg := git.GitClientConfig{
			Logger:   globals.AppLogger,
			Executor: globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)

			return fmt.Errorf("failed to initialize git client: %w", err)
		}

		var outputBuffer bytes.Buffer
		var hasChanges bool

		stagedOut, _, stagedErr := client.GetDiffCached(ctx)
		if stagedErr != nil {
			//nolint:wrapcheck // Wrapping is handled by caller.
			return stagedErr
		}
		if strings.TrimSpace(stagedOut) != "" {
			hasChanges = true
			tools.AppendSectionHeader(&outputBuffer, "Staged Changes")
			tools.AppendFencedCodeBlock(&outputBuffer, stagedOut, "diff")
		}

		unstagedOut, _, unstagedErr := client.GetDiffUnstaged(ctx)
		if unstagedErr != nil {
			//nolint:wrapcheck // Wrapping is handled by caller.
			return unstagedErr
		}
		if strings.TrimSpace(unstagedOut) != "" {
			hasChanges = true
			tools.AppendSectionHeader(&outputBuffer, "Unstaged Changes")
			tools.AppendFencedCodeBlock(&outputBuffer, unstagedOut, "diff")
		}

		untrackedOut, _, untrackedErr := client.ListUntrackedFiles(ctx)
		if untrackedErr != nil {
			//nolint:wrapcheck // Wrapping is handled by caller.
			return untrackedErr
		}
		if strings.TrimSpace(untrackedOut) != "" {
			hasChanges = true
			tools.AppendSectionHeader(&outputBuffer, "Untracked Files")
			tools.AppendFencedCodeBlock(&outputBuffer, untrackedOut, "")
		}

		presenter.Newline()
		if !hasChanges {
			presenter.Info("No pending changes found.")
		} else {
			errWrite := tools.WriteBufferToFile(config.DefaultDescribeOutputFile, &outputBuffer)
			if errWrite != nil {
				//nolint:wrapcheck // Wrapping is handled by caller.
				return errWrite
			}
		}

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(diffLongDescription, nil)
	if err != nil {
		panic(err)
	}

	DiffCmd.Short = desc.Short
	DiffCmd.Long = desc.Long
}
