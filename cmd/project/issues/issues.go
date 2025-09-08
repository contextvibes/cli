// cmd/project/issues/issues.go
package issues

import (
	"github.com/contextvibes/cli/cmd/project/issues/create"
	"github.com/contextvibes/cli/cmd/project/issues/list"
	"github.com/spf13/cobra"
)

// IssuesCmd represents the base command for the 'issues' subcommand group.
var IssuesCmd = &cobra.Command{
	Use:     "issues",
	Short:   "Manage project issues (work tickets, blueprints).",
	Aliases: []string{"issue"},
}

func init() {
	IssuesCmd.AddCommand(create.CreateCmd)
	IssuesCmd.AddCommand(list.ListCmd)
	// Other issue commands (link, export, etc.) would be added here
}
