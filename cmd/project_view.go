// cmd/project_view.go
package cmd

import (
	"context" // FIXED: Added missing import
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	issueAssignee    string
	issueLabel       string
	issueState       string
	issueSearchQuery string
	issueLimit       int
	viewAllHierarchy bool
)

// temp struct for unmarshalling 'gh issue view' JSON
type issueViewData struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	State  string   `json:"state"`
	Body   string   `json:"body"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

var projectViewCmd = &cobra.Command{
	Use:     "view [issue-number]",
	Aliases: []string{"list", "list-issues"},
	Short:   "Displays a list of issues or a detailed view of an issue's hierarchy.",
	Long: `Fetches and displays GitHub issues.

- With filters (--search, --label), it shows a flat list.
- With an issue number, it shows a detailed hierarchy for that single issue.
- With the --all flag, it shows a hierarchical view of ALL epics and their children.`,
	Example: `  # View all open issues
  contextvibes project view

  # View the hierarchy for Epic #24
  contextvibes project view 24
  
  # View the hierarchy for ALL epics
  contextvibes project view --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		if !ExecClient.CommandExists("gh") {
			presenter.Error("GitHub CLI ('gh') not found.")
			return errors.New("gh cli not found")
		}

		if viewAllHierarchy {
			return runAllHierarchyView(presenter, ctx)
		}
		if len(args) == 1 {
			return runSingleHierarchyView(presenter, ctx, args[0])
		}
		return runListView(presenter, ctx)
	},
}

func runListView(presenter *ui.Presenter, ctx context.Context) error {
	presenter.Summary("Fetching GitHub Issues...")
	ghArgs := []string{"issue", "list"}

	if issueSearchQuery != "" {
		ghArgs = append(ghArgs, "--search", issueSearchQuery)
		if issueLimit > 0 {
			ghArgs = append(ghArgs, "--limit", fmt.Sprintf("%d", issueLimit))
		}
	} else {
		if issueAssignee != "" {
			ghArgs = append(ghArgs, "--assignee", issueAssignee)
		}
		if issueLabel != "" {
			ghArgs = append(ghArgs, "--label", issueLabel)
		}
		if issueState != "" {
			ghArgs = append(ghArgs, "--state", issueState)
		}
		if issueLimit > 0 {
			ghArgs = append(ghArgs, "--limit", fmt.Sprintf("%d", issueLimit))
		}
	}

	if err := ExecClient.Execute(ctx, ".", "gh", ghArgs...); err != nil {
		presenter.Error("Failed to fetch GitHub issues: %v", err)
		return err
	}
	return nil
}

func runSingleHierarchyView(presenter *ui.Presenter, ctx context.Context, issueNumber string) error {
	presenter.Summary("Fetching hierarchy for Issue #%s...", issueNumber)
	viewArgs := []string{"issue", "view", issueNumber, "--json", "number,title,state,body,labels"}
	parentJSON, stderr, err := ExecClient.CaptureOutput(ctx, ".", "gh", viewArgs...)
	if err != nil {
		presenter.Error("Failed to fetch issue #%s: %v", issueNumber, err)
		presenter.Detail("Stderr: %s", stderr)
		return err
	}

	var parentData issueViewData
	if err := json.Unmarshal([]byte(parentJSON), &parentData); err != nil {
		presenter.Error("Failed to parse data for issue #%s: %v", issueNumber, err)
		return err
	}

	var labelNames []string
	for _, label := range parentData.Labels {
		labelNames = append(labelNames, label.Name)
	}
	presenter.Header("#%d: %s", parentData.Number, parentData.Title)
	presenter.Detail("State: %s | Labels: %s", parentData.State, strings.Join(labelNames, ", "))
	presenter.Newline()
	presenter.Step("Body:")
	_, _ = fmt.Fprintln(presenter.Out(), parentData.Body)
	presenter.Newline()

	re := regexp.MustCompile(`(?m)^\s*-\s\[( |x)\]\s#(\d+)`)
	matches := re.FindAllStringSubmatch(parentData.Body, -1)

	if len(matches) == 0 {
		presenter.Info("No child issues found in the task list.")
		return nil
	}

	presenter.Step("Child Issues:")
	for _, match := range matches {
		childNumber := match[2]
		childArgs := []string{"issue", "view", childNumber, "--json", "number,title,state"}
		childJSON, _, childErr := ExecClient.CaptureOutput(ctx, ".", "gh", childArgs...)
		if childErr != nil {
			presenter.Warning("  Could not fetch details for child #%s", childNumber)
			continue
		}

		var childData issueViewData
		if err := json.Unmarshal([]byte(childJSON), &childData); err != nil {
			presenter.Warning("  Could not parse details for child #%s", childNumber)
			continue
		}
		
		statusIcon := "✅"
		if childData.State == "OPEN" {
			statusIcon = "📝"
		}

		presenter.Detail("  %s #%d: %s (%s)", statusIcon, childData.Number, childData.Title, childData.State)
	}

	return nil
}

func runAllHierarchyView(presenter *ui.Presenter, ctx context.Context) error {
	presenter.Summary("Fetching Full Project Hierarchy...")

	presenter.Step("Finding all epics...")
	epicArgs := []string{"issue", "list", "--label", "epic", "--json", "number"}
	epicsJSON, stderr, err := ExecClient.CaptureOutput(ctx, ".", "gh", epicArgs...)
	if err != nil {
		presenter.Error("Failed to fetch epics: %v", err)
		presenter.Detail("Stderr: %s", stderr)
		return err
	}

	var epics []struct{ Number int `json:"number"` }
	if err := json.Unmarshal([]byte(epicsJSON), &epics); err != nil {
		presenter.Error("Failed to parse epic list: %v", err)
		return err
	}

	if len(epics) == 0 {
		presenter.Info("No epics found.")
		return nil
	}
	presenter.Success("✓ Found %d epics.", len(epics))
	presenter.Newline()

	for _, epic := range epics {
		if err := runSingleHierarchyView(presenter, ctx, fmt.Sprintf("%d", epic.Number)); err != nil {
			presenter.Warning("Could not display hierarchy for epic #%d. Continuing...", epic.Number)
		}
		presenter.Separator()
	}

	return nil
}

func init() {
	projectCmd.AddCommand(projectViewCmd)

	projectViewCmd.Flags().StringVarP(&issueAssignee, "assignee", "a", "", "Filter by assignee (@me to filter by yourself)")
	projectViewCmd.Flags().StringVarP(&issueLabel, "label", "l", "", "Filter by label")
	// FIXED: Changed .Flags. to .Flags().
	projectViewCmd.Flags().StringVarP(&issueState, "state", "s", "open", "Filter by state (open, closed, all)")
	projectViewCmd.Flags().IntVarP(&issueLimit, "limit", "L", 30, "Maximum number of issues to return")
	projectViewCmd.Flags().StringVar(&issueSearchQuery, "search", "", "Filter with a GitHub search query (e.g., 'is:open -label:bug')")
	projectViewCmd.Flags().BoolVar(&viewAllHierarchy, "all", false, "Display the full hierarchy of all epics and their children")
}
