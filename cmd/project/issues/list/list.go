// cmd/project/issues/list/list.go
package list

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed list.md.tpl
var listLongDescription string

var (
	issueAssignee    string
	issueLabel       string
	issueState       string
	issueSearchQuery string
	issueLimit       int
)

// ListCmd represents the project issues list command
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		execClient, ok := cmd.Context().Value("execClient").(*exec.ExecutorClient)
		if !ok { return errors.New("execClient not found in context") }
		
		presenter.Summary("Fetching GitHub Issues...")
		
		ghArgs := []string{"issue", "list"}
		if issueSearchQuery != "" {
			ghArgs = append(ghArgs, "--search", issueSearchQuery)
		}
		if issueAssignee != "" {
			ghArgs = append(ghArgs, "--assignee", issueAssignee)
		}
		if issueLabel != "" {
			ghArgs = append(ghArgs, "--label", issueLabel)
		}
		if issueState != "" {
			ghArgs = append(ghArgs, "--state", issueState)
		}
		if issueLimit > 0 {
			ghArgs = append(ghArgs, "--limit", fmt.Sprintf("%d", issueLimit))
		}

		return execClient.Execute(cmd.Context(), ".", "gh", ghArgs...)
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(listLongDescription, nil)
	if err != nil {
		panic(err)
	}
	ListCmd.Short = desc.Short
	ListCmd.Long = desc.Long

	ListCmd.Flags().StringVarP(&issueAssignee, "assignee", "a", "", "Filter by assignee")
	ListCmd.Flags().StringVarP(&issueLabel, "label", "l", "", "Filter by label")
	ListCmd.Flags().StringVarP(&issueState, "state", "s", "open", "Filter by state (open, closed, all)")
	ListCmd.Flags().IntVarP(&issueLimit, "limit", "L", 30, "Maximum number of issues to return")
	ListCmd.Flags().StringVar(&issueSearchQuery, "search", "", "Filter with a GitHub search query")
}
