package project
import (
	"github.com/contextvibes/cli/cmd/project/describe"
	"github.com/contextvibes/cli/cmd/project/issues"
	"github.com/spf13/cobra"
)
var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Commands for project planning and management (the 'why').",
}
func init() {
	ProjectCmd.AddCommand(describe.DescribeCmd)
	ProjectCmd.AddCommand(issues.IssuesCmd)
}
