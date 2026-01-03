// Package scaffold provides the command to generate infrastructure configuration.
package scaffold

import (
	_ "embed"
	"fmt"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workflow"
	"github.com/spf13/cobra"
)

//go:embed scaffold.md.tpl
var scaffoldLongDescription string

// ScaffoldCmd represents the scaffold command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ScaffoldCmd = &cobra.Command{
	Use:   "scaffold [target]",
	Short: "Scaffolds infrastructure (e.g., idx, firebase).",
	Example: `  contextvibes factory scaffold idx
  contextvibes factory scaffold firebase`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()
		target := args[0]

		runner := workflow.NewRunner(presenter, globals.AssumeYes)

		switch target {
		case "idx":
			return runner.Run(
				ctx,
				"Scaffolding IDX",
				&workflow.ScaffoldIDXStep{
					Presenter: presenter,
					AssumeYes: globals.AssumeYes,
				},
			)
		case "firebase":
			return runner.Run(
				ctx,
				"Scaffolding Firebase",
				&workflow.ScaffoldFirebaseStep{
					ExecClient: globals.ExecClient,
					Presenter:  presenter,
				},
			)
		default:
			//nolint:err113 // Dynamic error is appropriate for CLI output.
			return fmt.Errorf("unsupported scaffold target: %s (supported: idx, firebase)", target)
		}
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(scaffoldLongDescription, nil)
	if err != nil {
		panic(err)
	}

	ScaffoldCmd.Short = desc.Short
	ScaffoldCmd.Long = desc.Long
}
