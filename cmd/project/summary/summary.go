package summary

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"sync"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed summary.md.tpl
var summaryLongDescription string

const (
	concurrentFetches = 3
	maxBugsToList     = 5
	maxTasksToList    = 10
	maxEpicsToList    = 5
)

// SummaryCmd represents the project summary command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var SummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Displays a 'Morning Briefing' of project status.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}

		presenter.Summary("Project Morning Briefing")

		// We will fetch data in parallel
		var waitGroup sync.WaitGroup
		waitGroup.Add(concurrentFetches)

		var bugs, myTasks, epics []workitem.WorkItem
		var errBugs, errTasks, errEpics error

		// 1. Urgent Bugs
		go func() {
			defer waitGroup.Done()
			// Added 'is:issue' to fix 422 error
			bugs, errBugs = provider.SearchItems(ctx, "is:open is:issue label:bug sort:updated-desc")
		}()

		// 2. My Tasks (assignee:@me works in GitHub search)
		go func() {
			defer waitGroup.Done()
			// Added 'is:issue' to fix 422 error
			myTasks, errTasks = provider.SearchItems(ctx, "is:open is:issue assignee:@me sort:updated-desc")
		}()

		// 3. Active Epics
		go func() {
			defer waitGroup.Done()
			// Added 'is:issue' to fix 422 error
			epics, errEpics = provider.SearchItems(ctx, "is:open is:issue label:epic sort:updated-desc")
		}()

		presenter.Info("Fetching project data...")
		waitGroup.Wait()

		// --- Render: Urgent Attention ---
		presenter.Header("[!] Urgent Attention (Bugs)")
		if errBugs != nil {
			presenter.Warning("Could not fetch bugs: %v", errBugs)
		} else if len(bugs) == 0 {
			presenter.Success("No open bugs found. Great work!")
		} else {
			for _, item := range limit(bugs, maxBugsToList) {
				printItem(presenter, item)
			}
			if len(bugs) > maxBugsToList {
				presenter.Detail("... and %d more.", len(bugs)-maxBugsToList)
			}
		}
		presenter.Newline()

		// --- Render: On Your Plate ---
		presenter.Header("[@] On Your Plate (Assigned to You)")
		if errTasks != nil {
			presenter.Warning("Could not fetch your tasks: %v", errTasks)
		} else if len(myTasks) == 0 {
			presenter.Info("You have no assigned issues.")
		} else {
			for _, item := range limit(myTasks, maxTasksToList) {
				printItem(presenter, item)
			}
			if len(myTasks) > maxTasksToList {
				presenter.Detail("... and %d more.", len(myTasks)-maxTasksToList)
			}
		}
		presenter.Newline()

		// --- Render: Strategic Context ---
		presenter.Header("[#] Strategic Context (Active Epics)")
		if errEpics != nil {
			presenter.Warning("Could not fetch epics: %v", errEpics)
		} else if len(epics) == 0 {
			presenter.Info("No active epics found.")
		} else {
			for _, item := range limit(epics, maxEpicsToList) {

				_, _ = fmt.Fprintf(presenter.Out(), "  • #%d: %s\n", item.Number, item.Title)
			}
		}

		return nil
	},
}

func printItem(p *ui.Presenter, item workitem.WorkItem) {
	_, _ = fmt.Fprintf(p.Out(), "  - [#%d] %s\n", item.Number, item.Title)
}

func limit(items []workitem.WorkItem, limitCount int) []workitem.WorkItem {
	if len(items) > limitCount {
		return items[:limitCount]
	}

	return items
}

// newProvider is a factory function (duplicated from other cmds, ideally refactored later).
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

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(summaryLongDescription, nil)
	if err != nil {
		panic(err)
	}

	SummaryCmd.Short = desc.Short
	SummaryCmd.Long = desc.Long
}
