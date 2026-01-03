// Package list provides the command to list project issues.
package list

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/contextvibes/cli/cmd/project/issues/internal"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed list.md.tpl
var listLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	issueAssignee string
	issueLabel    string
	issueState    string
	issueLimit    int
	fullView      bool
)

// newProvider is a factory function that returns the configured work item provider.
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
		logger.DebugContext(
			ctx,
			"Work item provider not specified in config, defaulting to 'github'",
		)

		//nolint:wrapcheck // Factory function.
		return github.New(ctx, logger, cfg)
	default:
		//nolint:err113 // Dynamic error is appropriate here.
		return nil, fmt.Errorf(
			"unsupported work item provider '%s' specified in .contextvibes.yaml",
			cfg.Project.Provider,
		)
	}
}

// ListCmd represents the project issues list command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		//nolint:exhaustruct // Partial options are valid.
		listOpts := workitem.ListOptions{
			Limit:    issueLimit,
			Assignee: issueAssignee,
		}
		if issueLabel != "" {
			listOpts.Labels = []string{issueLabel}
		}
		switch issueState {
		case "closed":
			listOpts.State = workitem.StateClosed
		default:
			listOpts.State = workitem.StateOpen
		}

		presenter.Summary("Fetching Work Items...")
		items, err := provider.ListItems(ctx, listOpts)
		if err != nil {
			presenter.Error("Failed to list work items: %v", err)

			return fmt.Errorf("failed to list items: %w", err)
		}

		if len(items) == 0 {
			presenter.Info("No work items found matching the criteria.")

			return nil
		}

		if fullView {
			for _, item := range items {
				detailedItem, err := provider.GetItem(ctx, item.Number, false)
				if err != nil {
					presenter.Warning("Could not fetch details for #%d: %v", item.Number, err)

					continue
				}
				internal.DisplayWorkItem(presenter, detailedItem)
			}
		} else {
			for _, item := range items {

				fmt.Fprintf(presenter.Out(), "#%d [%s] %s\n", item.Number, item.State, item.Title)
			}
		}

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(listLongDescription, nil)
	if err != nil {
		panic(err)
	}

	ListCmd.Short = desc.Short
	ListCmd.Long = desc.Long

	ListCmd.Flags().StringVarP(&issueAssignee, "assignee", "a", "", "Filter by assignee")
	ListCmd.Flags().StringVarP(&issueLabel, "label", "l", "", "Filter by label")
	ListCmd.Flags().
		StringVarP(&issueState, "state", "s", "open", "Filter by state (open, closed, all)")
	//nolint:mnd // 30 is a reasonable default limit.
	ListCmd.Flags().IntVarP(&issueLimit, "limit", "L", 30, "Maximum number of issues to return")
	ListCmd.Flags().
		BoolVar(&fullView, "full", false, "Display the full details for each issue found")
}
