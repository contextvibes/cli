// cmd/project/issues/tree/tree.go
package tree

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/contextvibes/cli/cmd/project/issues/internal"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/contextvibes/cli/internal/workitem/resolver"
	"github.com/spf13/cobra"
)

//go:embed tree.md.tpl
var treeLongDescription string

var fullView bool

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

// TreeCmd represents the project issues tree command
var TreeCmd = &cobra.Command{
	Use:     "tree [issue-number]",
	Short:   "Display a hierarchical tree of epics, stories, and tasks.",
	Example: `  contextvibes project issues tree 52 --full`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)
			return err
		}

		resolver := resolver.New(provider)

		printFunc := printSummaryTree
		if fullView {
			printFunc = printFullTree
		}

		if len(args) > 0 {
			issueNumber, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid issue number provided: %s", args[0])
			}
			presenter.Summary("Building work item tree for Epic #%d...", issueNumber)
			root, err := resolver.BuildTree(ctx, issueNumber, fullView)
			if err != nil {
				presenter.Error("Failed to build work item tree: %v", err)
				return err
			}
			printFunc(presenter, root, 0)
		} else {
			presenter.Summary("Fetching all Epics to build trees...")
			listOpts := workitem.ListOptions{
				State:  workitem.StateOpen,
				Labels: []string{"epic"},
				Limit:  100,
			}
			epics, err := provider.ListItems(ctx, listOpts)
			if err != nil {
				presenter.Error("Failed to list epics: %v", err)
				return err
			}
			if len(epics) == 0 {
				presenter.Info("No open issues with the 'epic' label found.")
				return nil
			}

			for i, epic := range epics {
				root, err := resolver.BuildTree(ctx, epic.Number, fullView)
				if err != nil {
					presenter.Warning("Failed to build tree for Epic #%d: %v", epic.Number, err)
					continue
				}
				if i > 0 {
					presenter.Newline()
				}
				printFunc(presenter, root, 0)
			}
		}

		return nil
	},
}

// printSummaryTree recursively prints the work item hierarchy in a compact format.
func printSummaryTree(p *ui.Presenter, item *workitem.WorkItem, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(p.Out(), "%s- [%s] #%d: %s\n", indent, item.Type, item.Number, item.Title)
	for _, child := range item.Children {
		printSummaryTree(p, child, depth+1)
	}
}

// printFullTree recursively prints the work item hierarchy with full details.
func printFullTree(p *ui.Presenter, item *workitem.WorkItem, depth int) {
	indent := strings.Repeat("  ", depth)
	p.Out().Write([]byte(indent)) // Write indent manually for the header
	internal.DisplayWorkItem(p, item)

	if len(item.Comments) > 0 {
		fmt.Fprintf(p.Out(), "%s--- Comments (%d) ---\n", indent, len(item.Comments))
		for _, comment := range item.Comments {
			p.Out().Write([]byte(indent))
			p.Header(
				fmt.Sprintf(
					"Comment by %s on %s",
					comment.Author,
					comment.CreatedAt.Format("2006-01-02"),
				),
			)
			// Indent the body of the comment
			for _, line := range strings.Split(comment.Body, "\n") {
				fmt.Fprintf(p.Out(), "%s  %s\n", indent, line)
			}
			p.Out().Write([]byte(indent))
			p.Separator()
		}
	}

	for _, child := range item.Children {
		p.Newline()
		printFullTree(p, child, depth+1)
	}
}

func init() {
	desc, err := cmddocs.ParseAndExecute(treeLongDescription, nil)
	if err != nil {
		panic(err)
	}
	TreeCmd.Long = desc.Long
	TreeCmd.Flags().
		BoolVar(&fullView, "full", false, "Display the full details, including body and comments, for each issue in the tree.")
}
