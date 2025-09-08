// cmd/project/board/list/list.go
package list

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed list.md.tpl
var listLongDescription string

// newGHClient is a factory function that returns a configured GitHub client.
func newGHClient(ctx context.Context, logger *slog.Logger, cfg *config.Config) (*github.Client, error) {
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
	// Use the shared parser from the internal/github package
	owner, repo, err := github.ParseGitHubRemote(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse owner/repo from remote URL '%s': %w", remoteURL, err)
	}
	return github.NewClient(ctx, logger, owner, repo)
}

// ListCmd represents the project board list command
var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists available project boards.",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		ghClient, err := newGHClient(ctx, globals.AppLogger, globals.LoadedAppConfig)
		if err != nil {
			presenter.Error("Failed to initialize GitHub client: %v", err)
			return err
		}

		presenter.Summary("Fetching project boards...")
		projects, err := ghClient.ListProjects(ctx)
		if err != nil {
			presenter.Error("Failed to fetch project boards: %v", err)
			presenter.Advice("Please ensure your GITHUB_TOKEN has the 'read:project' scope.")
			return err
		}

		if len(projects) == 0 {
			presenter.Info("No project boards found for this repository's owner.")
			return nil
		}

		presenter.Header("--- Available Project Boards ---")
		for _, project := range projects {
			presenter.Step("#%d: %s", project.Number, project.Title)
			presenter.Detail(project.URL)
		}

		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(listLongDescription, nil)
	if err != nil {
		panic(err)
	}
	ListCmd.Long = desc.Long
}
