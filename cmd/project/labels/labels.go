// cmd/project/labels/labels.go
package labels

import (
	"github.com/contextvibes/cli/cmd/project/labels/create"
	"github.com/spf13/cobra"
)

// LabelsCmd represents the base command for the 'labels' subcommand group.
var LabelsCmd = &cobra.Command{
	Use:     "labels",
	Short:   "Manage project labels.",
	Aliases: []string{"label"},
}

func init() {
	LabelsCmd.AddCommand(create.CreateCmd)
}
