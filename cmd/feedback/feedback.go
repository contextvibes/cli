// Package feedback provides the command to submit feedback.
package feedback

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	wigh "github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed feedback.md.tpl
var feedbackLongDescription string

// newProviderForRepo creates a workitem.Provider for a specific owner/repo string.
//
//nolint:ireturn // Returning interface is intended for provider abstraction.
func newProviderForRepo(
	ctx context.Context,
	logger *slog.Logger,
	owner, repo string,
) (workitem.Provider, error) {
	ghClient, err := github.NewClient(ctx, logger, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create github api client for %s/%s: %w", owner, repo, err)
	}

	return wigh.NewWithClient(ghClient, logger, owner, repo), nil
}

// FeedbackCmd represents the feedback command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var FeedbackCmd = &cobra.Command{
	Use:     "feedback [repo-alias] [title]",
	Short:   "Submit feedback to a contextvibes repository.",
	Example: `  contextvibes feedback "Tree command is slow"`,
	//nolint:mnd // 2 arguments max.
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()
		cfg := globals.LoadedAppConfig.Feedback

		var repoAlias, title, body string
		repoAlias = cfg.DefaultRepository

		if len(args) == 1 {
			if _, ok := cfg.Repositories[args[0]]; ok {
				repoAlias = args[0]
			} else {
				title = args[0]
			}
		} else if len(args) == 2 { //nolint:mnd // 2 arguments check.
			repoAlias = args[0]
			title = args[1]
		}

		targetRepo, ok := cfg.Repositories[repoAlias]
		if !ok {
			//nolint:err113 // Dynamic error is appropriate here.
			return fmt.Errorf("repository alias '%s' not found in configuration", repoAlias)
		}
		repoParts := strings.Split(targetRepo, "/")
		//nolint:mnd // Expecting owner/repo.
		if len(repoParts) != 2 {
			//nolint:err113 // Dynamic error is appropriate here.
			return fmt.Errorf(
				"invalid repository format for alias '%s': expected 'owner/repo', got '%s'",
				repoAlias,
				targetRepo,
			)
		}
		owner, repo := repoParts[0], repoParts[1]

		if title == "" {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("What is the title of your feedback?").Value(&title),
					huh.NewText().Title("Please provide more details (optional)").Value(&body),
				),
			)
			err := form.Run()
			if err != nil {
				return fmt.Errorf("input form failed: %w", err)
			}
		}
		if strings.TrimSpace(title) == "" {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("title cannot be empty")
		}

		provider, err := newProviderForRepo(ctx, globals.AppLogger, owner, repo)
		if err != nil {
			presenter.Error("Failed to initialize provider for %s: %v", targetRepo, err)

			return err
		}

		ghClient, _ := github.NewClient(ctx, globals.AppLogger, owner, repo)
		user, _ := ghClient.GetAuthenticatedUserLogin(ctx)
		if user == "" {
			user = "unknown"
		}

		contextBlock := fmt.Sprintf(
			"\n\n---\n**Context**\n- **CLI Version:** `%s`\n- **OS/Arch:** `%s/%s`\n- **Filed by:** @%s",
			globals.AppVersion,
			runtime.GOOS,
			runtime.GOARCH,
			user,
		)
		finalBody := body + contextBlock

		//nolint:exhaustruct // Partial initialization is valid for creation.
		newItem := workitem.WorkItem{
			Title:  title,
			Body:   finalBody,
			Labels: []string{"feedback"},
		}

		presenter.Summary("Submitting feedback to %s...", targetRepo)
		createdItem, err := provider.CreateItem(ctx, newItem)
		if err != nil {
			presenter.Error("Failed to create issue: %v", err)
			presenter.Advice(
				"Please ensure your GITHUB_TOKEN has the 'repo' scope for the '%s' repository.",
				targetRepo,
			)

			return fmt.Errorf("failed to create item: %w", err)
		}

		presenter.Success("✓ Thank you! Your feedback has been submitted: %s", createdItem.URL)

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(feedbackLongDescription, nil)
	if err != nil {
		panic(err)
	}

	FeedbackCmd.Long = desc.Long
}
