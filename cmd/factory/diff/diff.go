// cmd/factory/diff/diff.go
package diff

import (
	"bytes"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed diff.md.tpl
var diffLongDescription string

const fixedDiffOutputFile = "contextvibes.md"

// DiffCmd represents the diff command
var DiffCmd = &cobra.Command{
	Use:     "diff",
	Example: `  contextvibes factory diff`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		
		logger, ok := cmd.Context().Value("logger").(*slog.Logger)
		if !ok { return errors.New("logger not found in context") }
		ctx := cmd.Context()

		presenter.Summary("Generating Git diff summary for %s.", fixedDiffOutputFile)

		workDir, err := os.Getwd()
		if err != nil {
			return err
		}
		
		gitCfg := git.GitClientConfig{Logger: logger}
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}

		var outputBuffer bytes.Buffer
		var hasChanges bool

		stagedOut, _, stagedErr := client.GetDiffCached(ctx)
		if stagedErr != nil { return stagedErr }
		if strings.TrimSpace(stagedOut) != "" {
			hasChanges = true
			tools.AppendSectionHeader(&outputBuffer, "Staged Changes")
			tools.AppendFencedCodeBlock(&outputBuffer, stagedOut, "diff")
		}

		unstagedOut, _, unstagedErr := client.GetDiffUnstaged(ctx)
		if unstagedErr != nil { return unstagedErr }
		if strings.TrimSpace(unstagedOut) != "" {
			hasChanges = true
			tools.AppendSectionHeader(&outputBuffer, "Unstaged Changes")
			tools.AppendFencedCodeBlock(&outputBuffer, unstagedOut, "diff")
		}

		untrackedOut, _, untrackedErr := client.ListUntrackedFiles(ctx)
		if untrackedErr != nil { return untrackedErr }
		if strings.TrimSpace(untrackedOut) != "" {
			hasChanges = true
			tools.AppendSectionHeader(&outputBuffer, "Untracked Files")
			tools.AppendFencedCodeBlock(&outputBuffer, untrackedOut, "")
		}

		presenter.Newline()
		if !hasChanges {
			presenter.Info("No pending changes found.")
		} else {
			if errWrite := tools.WriteBufferToFile(fixedDiffOutputFile, &outputBuffer); errWrite != nil {
				return errWrite
			}
		}
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(diffLongDescription, nil)
	if err != nil {
		panic(err)
	}
	DiffCmd.Short = desc.Short
	DiffCmd.Long = desc.Long
}
