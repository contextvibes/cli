// cmd/project/labels/create/create.go
package create

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed create.md.tpl
var createLongDescription string

var (
	labelName        string
	labelDescription string
	labelColor       string
)

// newProvider is a factory function that returns the configured work item provider.
func newProvider(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (workitem.Provider, error) {
	switch cfg.Project.Provider {
	case "github":
		return github.New(ctx, logger, cfg)
	case "":
		logger.DebugContext(
			ctx,
			"Work item provider not specified in config, defaulting to 'github'",
		)

		return github.New(ctx, logger, cfg)
	default:
		return nil, fmt.Errorf(
			"unsupported work item provider '%s' specified in .contextvibes.yaml",
			cfg.Project.Provider,
		)
	}
}

// CreateCmd represents the project labels create command.
var CreateCmd = &cobra.Command{
	Use:     "create --name <label-name>",
	Short:   "Create a new label in the repository.",
	Example: `  contextvibes project labels create --name "docs" --description "Documentation updates" --color "0075ca"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if strings.TrimSpace(labelName) == "" {
			return errors.New("label name cannot be empty, please provide the --name flag")
		}

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		newLabel := workitem.Label{
			Name:        labelName,
			Description: labelDescription,
			Color:       labelColor,
		}

		presenter.Summary("Creating label '%s'...", newLabel.Name)
		_, err = provider.CreateLabel(ctx, newLabel)
		if err != nil {
			presenter.Error("Failed to create label: %v", err)

			return err
		}

		presenter.Success("Successfully created label '%s'.", newLabel.Name)

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(createLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CreateCmd.Long = desc.Long

	CreateCmd.Flags().StringVarP(&labelName, "name", "n", "", "The name of the label (required)")
	CreateCmd.Flags().
		StringVarP(&labelDescription, "description", "d", "", "A description for the label")
	CreateCmd.Flags().
		StringVarP(&labelColor, "color", "c", "", "A 6-character hex color code for the label (without the #)")

	if err := CreateCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
}
