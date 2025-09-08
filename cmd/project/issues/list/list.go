// cmd/project/issues/list/list.go
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

var (
	issueAssignee string
	issueLabel    string
	issueState    string
	issueLimit    int
	fullView      bool
)

// newProvider is a factory function that returns the configured work item provider.
func newProvider(ctx context.Context, logger *slog.Logger, cfg *config.Config) (workitem.Provider, error) {
	switch cfg.Project.Provider {
	case "github":
		return github.New(ctx, logger, cfg)
	case "":
		logger.DebugContext(ctx, "Work item provider not specified in config, defaulting to 'github'")
		return github.New(ctx, logger, cfg)
	default:
		return nil, fmt.Errorf("unsupported work item provider '%s' specified in .contextvibes.yaml", cfg.Project.Provider)
	}
}

// ListCmd represents the project issues list command
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)
			return err
		}

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
			return err
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

func init() {
	desc, err := cmddocs.ParseAndExecute(listLongDescription, nil)
	if err != nil {
		panic(err)
	}
	ListCmd.Short = desc.Short
	ListCmd.Long = desc.Long

	ListCmd.Flags().StringVarP(&issueAssignee, "assignee", "a", "", "Filter by assignee")
	ListCmd.Flags().StringVarP(&issueLabel, "label", "l", "", "Filter by label")
	ListCmd.Flags().StringVarP(&issueState, "state", "s", "open", "Filter by state (open, closed, all)")
	ListCmd.Flags().IntVarP(&issueLimit, "limit", "L", 30, "Maximum number of issues to return")
	ListCmd.Flags().BoolVar(&fullView, "full", false, "Display the full details for each issue found")
}
