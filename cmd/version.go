package cmd

import (
	_ "embed"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed version.md.tpl
var versionLongDescription string

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use: "version",
	// Short and Long descriptions are now set dynamically in the init() function.
	RunE: func(cmd *cobra.Command, args []string) error {
		p := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		p.Summary("Context Vibes CLI Version: " + AppVersion)
		return nil
	},
}

func init() {
	// Dynamically set the descriptions by parsing and executing the template.
	desc, err := cmddocs.ParseAndExecute(
		versionLongDescription,
		struct{ AppVersion string }{AppVersion: AppVersion},
	)
	if err != nil {
		// This is a developer error, so we panic.
		panic(err)
	}
	versionCmd.Short = desc.Short
	versionCmd.Long = desc.Long

	rootCmd.AddCommand(versionCmd)
}
