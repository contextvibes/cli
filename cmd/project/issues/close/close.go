package closecmd

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed close.md.tpl
var closeLongDescription string

// newProvider is a factory function (duplicated from other cmds).
//

func newProvider(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (workitem.Provider, error) {
	switch cfg.Project.Provider {
	case "github":
		//nolint:wrapcheck // Factory function.
		return github.New(ctx, logger, cfg)
	case "":
		//nolint:wrapcheck // Factory function.
		return github.New(ctx, logger, cfg)
	default:
		//nolint:err113 // Dynamic error is appropriate here.
		return nil, fmt.Errorf(
			"unsupported work item provider '%s'",
			cfg.Project.Provider,
		)
	}
}

// CloseCmd represents the project issues close command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var CloseCmd = &cobra.Command{
	Use:     "close <issue-number>",
	Short:   "Closes a specific issue.",
	Example: "  contextvibes project issues close 29",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		issueNumber, err := strconv.Atoi(args[0])
		if err != nil {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("invalid issue number provided")
		}

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		// 1. Fetch the item first to ensure it exists and get current data
		// This prevents overwriting other fields with empty values if the provider isn't careful
		presenter.Step("Fetching issue #%d...", issueNumber)
		item, err := provider.GetItem(ctx, issueNumber, false)
		if err != nil {
			presenter.Error("Failed to fetch issue: %v", err)

			return fmt.Errorf("failed to get item: %w", err)
		}

		if item.State == workitem.StateClosed {
			presenter.Info("Issue #%d is already closed.", issueNumber)

			return nil
		}

		// 2. Update state locally
		item.State = workitem.StateClosed

		// 3. Send update
		presenter.Step("Closing issue #%d...", issueNumber)
		_, err = provider.UpdateItem(ctx, issueNumber, *item)
		if err != nil {
			presenter.Error("Failed to close issue: %v", err)

			return fmt.Errorf("failed to update item: %w", err)
		}

		presenter.Success("✓ Issue #%d closed successfully.", issueNumber)

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(closeLongDescription, nil)
	if err != nil {
		panic(err)
	}

	CloseCmd.Short = desc.Short
	CloseCmd.Long = desc.Long
}
