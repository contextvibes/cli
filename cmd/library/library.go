package library
import (
	"github.com/contextvibes/cli/cmd/library/index"
	"github.com/contextvibes/cli/cmd/library/systemprompt"
	"github.com/contextvibes/cli/cmd/library/thea"
	"github.com/spf13/cobra"
)
var LibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Commands for knowledge and standards (the 'where').",
}
func init() {
	LibraryCmd.AddCommand(index.IndexCmd)
	LibraryCmd.AddCommand(thea.TheaCmd)
	LibraryCmd.AddCommand(systemprompt.SystemPromptCmd)
}
