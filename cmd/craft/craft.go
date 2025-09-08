package craft
import (
	"github.com/contextvibes/cli/cmd/craft/kickoff"
	"github.com/contextvibes/cli/cmd/craft/message"
	"github.com/contextvibes/cli/cmd/craft/prdescription"
	"github.com/spf13/cobra"
)
var CraftCmd = &cobra.Command{
	Use:   "craft",
	Short: "Commands for AI-assisted creative tasks (the 'who' & 'how').",
}
func init() {
	CraftCmd.AddCommand(message.MessageCmd)
	CraftCmd.AddCommand(prdescription.PRDescriptionCmd)
	CraftCmd.AddCommand(kickoff.KickoffCmd)
}
