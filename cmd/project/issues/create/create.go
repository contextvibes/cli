// cmd/project/issues/create/create.go
package create

import (
	_ "embed"
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed create.md.tpl
var createLongDescription string

var (
	issueType         string
	issueTitle        string
	issueBody         string
	parentIssueNumber int
)

// CreateCmd represents the project issues create command
var CreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"new", "add"},
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		execClient, ok := cmd.Context().Value("execClient").(*exec.ExecutorClient)
		if !ok { return errors.New("execClient not found in context") }
		assumeYes, ok := cmd.Context().Value("assumeYes").(bool)
		if !ok { return errors.New("assumeYes not found in context") }
		ctx := cmd.Context()

		if !execClient.CommandExists("gh") {
			presenter.Error("GitHub CLI ('gh') not found.")
			return errors.New("gh cli not found")
		}

		if issueTitle == "" { // Interactive Mode
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().Title("What kind of issue is this?").Options(huh.NewOption("Feature", "feature"), huh.NewOption("Bug", "bug"), huh.NewOption("Chore", "chore")).Value(&issueType),
					huh.NewInput().Title("Title?").Value(&issueTitle),
					huh.NewText().Title("Body?").Value(&issueBody),
				),
			)
			if err := form.Run(); err != nil { return err }
		}

		if !assumeYes {
			// A confirmation step would go here
		}

		ghArgs := []string{"issue", "create", "--title", issueTitle, "--body", issueBody, "--label", issueType}
		stdout, _, err := execClient.CaptureOutput(ctx, ".", "gh", ghArgs...)
		if err != nil {
			presenter.Error("Failed to create GitHub issue: %v", err)
			return err
		}

		presenter.Success("Successfully created issue: %s", strings.TrimSpace(stdout))
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(createLongDescription, nil)
	if err != nil {
		panic(err)
	}
	CreateCmd.Short = desc.Short
	CreateCmd.Long = desc.Long

	CreateCmd.Flags().StringVarP(&issueType, "type", "t", "", "Type of the issue (feature, bug, chore)")
	CreateCmd.Flags().StringVarP(&issueTitle, "title", "T", "", "Title of the issue")
	CreateCmd.Flags().StringVarP(&issueBody, "body", "b", "", "Body of the issue")
	CreateCmd.Flags().IntVarP(&parentIssueNumber, "parent", "p", 0, "Parent issue number")
}
