// cmd/project/board/add/add.go
package add

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	wigh "github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed add.md.tpl
var addLongDescription string

// newGHClient is a factory function that returns a configured GitHub client.
func newGHClient(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (*github.Client, error) {
	gitClient, err := git.NewClient(ctx, ".", git.GitClientConfig{
		Executor: globals.ExecClient.UnderlyingExecutor(),
		Logger:   logger,
	})
	if err != nil {
		return nil, fmt.Errorf("could not initialize git client for repo discovery: %w", err)
	}

	remoteURL, err := gitClient.GetRemoteURL(ctx, cfg.Git.DefaultRemote)
	if err != nil {
		return nil, fmt.Errorf("could not get remote URL for '%s': %w", cfg.Git.DefaultRemote, err)
	}

	owner, repo, err := github.ParseGitHubRemote(remoteURL)
	if err != nil {
		return nil, fmt.Errorf(
			"could not parse owner/repo from remote URL '%s': %w",
			remoteURL,
			err,
		)
	}

	return github.NewClient(ctx, logger, owner, repo)
}

// newProvider is a factory function that returns the configured work item provider.
func newProvider(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (workitem.Provider, error) {
	return wigh.New(ctx, logger, cfg)
}

// AddCmd represents the project board add command.
var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactively add issues to a project board.",
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		ghClient, err := newGHClient(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize GitHub client: %v", err)

			return err
		}

		// 1. Select a Project Board
		presenter.Summary("Step 1: Select a Project Board")
		projects, err := ghClient.ListProjects(ctx)
		if err != nil {
			presenter.Error("Failed to fetch project boards: %v", err)
			presenter.Advice(
				"Please ensure your GITHUB_TOKEN has the 'read:project' and 'write:project' scopes.",
			)

			return err
		}
		if len(projects) == 0 {
			presenter.Info("No project boards found.")

			return nil
		}
		projectOptions := make([]string, len(projects))
		for i, p := range projects {
			projectOptions[i] = fmt.Sprintf("#%d: %s", p.Number, p.Title)
		}
		selectedProjectStr, err := presenter.PromptForSelect(
			"Which project board do you want to add issues to?",
			projectOptions,
		)
		if err != nil {
			return err // User likely cancelled
		}
		projectNumberStr := strings.Split(selectedProjectStr, ":")[0][1:]
		projectNumber, _ := strconv.Atoi(projectNumberStr)
		project, err := ghClient.GetProjectByNumber(ctx, projectNumber)
		if err != nil {
			presenter.Error("Failed to get details for project #%d: %v", projectNumber, err)

			return err
		}
		presenter.Success("✓ Selected board '%s'", project.Title)
		presenter.Newline()

		// 2. Select Issues to Add
		presenter.Summary("Step 2: Select Issues to Add")
		provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize work item provider: %v", err)

			return err
		}
		listOpts := workitem.ListOptions{State: workitem.StateOpen, Limit: 100}
		issues, err := provider.ListItems(ctx, listOpts)
		if err != nil {
			presenter.Error("Failed to list open issues: %v", err)

			return err
		}
		if len(issues) == 0 {
			presenter.Info("No open issues found to add.")

			return nil
		}
		issueOptions := make([]string, len(issues))
		for i, issue := range issues {
			issueOptions[i] = fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
		}
		selectedIssueStrs, err := presenter.PromptForMultiSelect(
			"Which issues would you like to add?",
			issueOptions,
		)
		if err != nil {
			return err // User likely cancelled
		}
		if len(selectedIssueStrs) == 0 {
			presenter.Info("No issues selected. Aborting.")

			return nil
		}
		presenter.Newline()

		// 3. Add Issues
		presenter.Summary(
			"Step 3: Adding %d Issue(s) to '%s'",
			len(selectedIssueStrs),
			project.Title,
		)
		for _, issueStr := range selectedIssueStrs {
			issueNumberStr := strings.Split(issueStr, ":")[0][1:]
			issueNumber, _ := strconv.Atoi(issueNumberStr)

			presenter.Step("Adding issue #%d...", issueNumber)
			item, err := provider.GetItem(ctx, issueNumber, false)
			if err != nil {
				presenter.Error("  ! Failed to get details for issue #%d: %v", issueNumber, err)

				continue
			}
			if item.ID == "" {
				presenter.Error("  ! Could not find GraphQL Node ID for issue #%d.", issueNumber)

				continue
			}

			err = ghClient.AddIssueToProject(ctx, project.ID, item.ID)
			if err != nil {
				presenter.Error("  ! Failed to add issue #%d to project: %v", issueNumber, err)
			} else {
				presenter.Success("  ✓ Successfully added issue #%d.", issueNumber)
			}
		}

		presenter.Newline()
		presenter.Success("Finished adding issues. View the board at: %s", project.URL)

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(addLongDescription, nil)
	if err != nil {
		panic(err)
	}

	AddCmd.Long = desc.Long
}
