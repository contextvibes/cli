// cmd/project/board/board.go
package board

import (
	"github.com/contextvibes/cli/cmd/project/board/add"
	"github.com/contextvibes/cli/cmd/project/board/list"
	"github.com/spf13/cobra"
)

// BoardCmd represents the base command for the 'board' subcommand group.
var BoardCmd = &cobra.Command{
	Use:     "board",
	Short:   "Manage project boards.",
	Aliases: []string{"boards"},
}

func init() {
	BoardCmd.AddCommand(list.ListCmd)
	BoardCmd.AddCommand(add.AddCmd)
}
