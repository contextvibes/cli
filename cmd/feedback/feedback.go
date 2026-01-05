package feedback

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/build"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	wigh "github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

const maxArgs = 2

var (
	// ErrRepoAliasNotFound is returned when a repository alias is not in the config.
	ErrRepoAliasNotFound = errors.New("repository alias not found in configuration")
	// ErrInvalidRepoFormat is returned when a repository format is invalid.
	ErrInvalidRepoFormat = errors.New("invalid repository format in configuration")
	// ErrTitleEmpty is returned when a user provides an empty title.
	ErrTitleEmpty = errors.New("title cannot be empty")
)

//go:embed feedback.md.tpl
var feedbackLongDescription string

// feedbackParams holds the parameters for the feedback command logic.
type feedbackParams struct {
	presenter *ui.Presenter
	logger    *slog.Logger
	cfg       *config.FeedbackSettings
	repoAlias string
	title     string
	body      string
}

// NewFeedbackCmd creates and configures the `feedback` command.
func NewFeedbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "feedback [repo-alias] [title]",
		Short:   "Submit feedback to a contextvibes repository.",
		Example: "  contextvibes feedback \"Tree command is slow\"",
		Args:    cobra.MaximumNArgs(maxArgs),
		RunE:    runFeedbackCmd,
	}

	desc, err := cmddocs.ParseAndExecute(feedbackLongDescription, nil)
	if err == nil {
		cmd.Long = desc.Long
	}

	return cmd
}

// runFeedbackCmd is the main execution logic for the feedback command.
func runFeedbackCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	params := &feedbackParams{
		presenter: ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr()),
		logger:    globals.AppLogger,
		cfg:       &globals.LoadedAppConfig.Feedback,
		repoAlias: "",
		title:     "",
		body:      "",
	}

	// 1. Parse arguments & Resolve Target
	owner, repo, err := resolveTarget(args, params)
	if err != nil {
		return err
	}

	// 2. Gather user input if needed (Interactive)
	if err := ensureFeedbackContent(params); err != nil {
		return err
	}

	// 3. Initialize GitHub Client (Network)
	ghClient, err := github.NewClient(ctx, params.logger, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to create github client: %w", err)
	}

	// 4. Fetch User Identity (Network)
	user, err := ghClient.GetAuthenticatedUserLogin(ctx)
	if err != nil {
		params.logger.Warn("Could not determine authenticated user", "error", err)
		user = "unknown"
	}

	// 5. Submit
	provider := wigh.NewWithClient(ghClient, params.logger, owner, repo)
	newItem := constructWorkItem(params, user, build.Version)

	return submitFeedback(ctx, params.presenter, provider, newItem, fmt.Sprintf("%s/%s", owner, repo))
}

// resolveTarget determines the owner and repo based on args and config.
func resolveTarget(args []string, params *feedbackParams) (string, string, error) {
	alias, title := parseFeedbackArgs(args, params.cfg.DefaultRepository, params.cfg.Repositories)
	params.repoAlias = alias
	params.title = title

	targetRepo, ok := params.cfg.Repositories[params.repoAlias]
	if !ok {
		return "", "", fmt.Errorf("%w: %s", ErrRepoAliasNotFound, params.repoAlias)
	}

	repoParts := strings.Split(targetRepo, "/")
	if len(repoParts) != maxArgs {
		return "", "", fmt.Errorf("%w: expected 'owner/repo', got '%s'", ErrInvalidRepoFormat, targetRepo)
	}

	return repoParts[0], repoParts[1], nil
}

// parseFeedbackArgs determines the repository alias and title from command-line arguments.
func parseFeedbackArgs(args []string, defaultRepo string, repositories map[string]string) (string, string) {
	repoAlias := defaultRepo
	var title string

	if len(args) > 0 {
		// Check if the first argument matches a known repository alias
		if _, ok := repositories[args[0]]; ok {
			repoAlias = args[0]
			if len(args) > 1 {
				title = args[1]
			}
		} else {
			// If not an alias, assume it's the title
			title = args[0]
		}
	}

	return repoAlias, title
}

// ensureFeedbackContent prompts the user for a title and body if not already provided.
func ensureFeedbackContent(params *feedbackParams) error {
	if params.title == "" {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("What is the title of your feedback?").Value(&params.title),
				huh.NewText().Title("Please provide more details (optional)").Value(&params.body),
			),
		)
		if err := form.Run(); err != nil {
			return fmt.Errorf("input form failed: %w", err)
		}
	}

	if strings.TrimSpace(params.title) == "" {
		return ErrTitleEmpty
	}

	return nil
}

// constructWorkItem builds the WorkItem struct for submission.
func constructWorkItem(params *feedbackParams, user, appVersion string) workitem.WorkItem {
	contextBlock := fmt.Sprintf(
		"\n\n---\n**Context**\n- **CLI Version:** `%s`\n- **OS/Arch:** `%s/%s`\n- **Filed by:** @%s",
		appVersion,
		runtime.GOOS,
		runtime.GOARCH,
		user,
	)

	finalBody := params.body + contextBlock

	return workitem.WorkItem{
		ID:        "",
		Number:    0,
		State:     "",
		Type:      "issue",
		Title:     params.title,
		Body:      finalBody,
		URL:       "",
		Author:    user,
		Labels:    []string{"feedback"},
		Assignees: nil,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
		Comments:  nil,
		Children:  nil,
	}
}

// submitFeedback creates the item using the provider and reports the result.
func submitFeedback(
	ctx context.Context,
	presenter *ui.Presenter,
	provider workitem.Provider,
	item workitem.WorkItem,
	targetRepo string,
) error {
	presenter.Summary("Submitting feedback to %s...", targetRepo)

	createdItem, err := provider.CreateItem(ctx, item)
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
}
