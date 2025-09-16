// cmd/craft/message/message.go
package message

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed message.md.tpl
var messageLongDescription string

// MessageCmd represents the craft message command
var MessageCmd = &cobra.Command{
	Use:     "message",
	Aliases: []string{"commit", "msg"},
	Short:   "Generates a suggested 'factory commit' command.",
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Crafting a commit message...")

		stagedDiff, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "git", "diff", "--staged")
		if err != nil {
			return err
		}

		if strings.TrimSpace(stagedDiff) == "" {
			presenter.Info("No staged changes found to generate a commit message from.")
			presenter.Advice("Please stage your changes using 'git add' first.")
			return nil
		}

		// In a real implementation, this is where the call to an LLM would happen.
		// We would send the 'stagedDiff' and get back a structured response.
		// For now, we will simulate this with a placeholder.

		presenter.Info("AI analysis complete. Suggested command:")
		presenter.Newline()

		// Simulated AI response:
		simulatedSubject := "feat(craft): add placeholder for message generation"
		simulatedBody := "This change introduces the 'craft message' command but uses a hardcoded placeholder for the AI-generated commit message. The real implementation will call an LLM."

		fmt.Fprintf(
			presenter.Out(),
			"contextvibes factory commit -m \"%s\" -m \"%s\"\n",
			simulatedSubject,
			simulatedBody,
		)

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(messageLongDescription, nil)
	if err != nil {
		panic(err)
	}
	MessageCmd.Short = desc.Short
	MessageCmd.Long = desc.Long
}
