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
	"time"

	"github.com/charmbracelet/huh"
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
		Example: `  contextvibes feedback "Tree command is slow"`,
		Args:    cobra.MaximumNArgs(maxArgs),
		RunE:    runFeedbackCmd,

		// Boilerplate
		GroupID:                    "",
		Long:                       "", // Will be set from embedded doc
		Aliases:                    []string{},
		SuggestFor:                 []string{},
		ValidArgs:                  []string{},
		ValidArgsFunction:          nil,
		ArgAliases:                 []string{},
		BashCompletionFunction:     "",
		Deprecated:                 "",
		Annotations:                nil,
		Version:                    "",
		PersistentPreRun:           nil,
		PersistentPreRunE:          nil,
		PreRun:                     nil,
		PreRunE:                    nil,
		PostRun:                    nil,
		PostRunE:                   nil,
		PersistentPostRun:          nil,
		PersistentPostRunE:         nil,
		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
		CompletionOptions:          cobra.CompletionOptions{},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              true,
		SilenceUsage:               true,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          true,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}

	desc, err := cmddocs.ParseAndExecute(feedbackLongDescription, nil)
	if err == nil {
		cmd.Long = desc.Long
	}

	return cmd
}

// runFeedbackCmd is the main execution logic for the feedback command.
func runFeedbackCmd(cmd *cobra.Command, args []string) error {
	params := &feedbackParams{
		presenter: ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr()),
		logger:    globals.AppLogger,
		cfg:       &globals.LoadedAppConfig.Feedback,
		repoAlias: "",
		title:     "",
		body:      "",
	}

	// 1. Parse arguments
	alias, title := parseFeedbackArgs(args, params.cfg.DefaultRepository, params.cfg.Repositories)
	params.repoAlias = alias
	params.title = title

	// 2. Get target repository and validate
	targetRepo, ok := params.cfg.Repositories[params.repoAlias]
	if !ok {
		return fmt.Errorf("%w: %s", ErrRepoAliasNotFound, params.repoAlias)
	}

	repoParts := strings.Split(targetRepo, "/")
	if len(repoParts) != maxArgs {
		return fmt.Errorf("%w: expected 'owner/repo', got '%s'", ErrInvalidRepoFormat, targetRepo)
	}

	owner, repo := repoParts[0], repoParts[1]

	// 3. Gather user input if needed
	if err := gatherFeedbackFromUser(&params.title, &params.body); err != nil {
		return err
	}

	// 4. Create provider and GitHub client
	provider, err := newProviderForRepo(cmd.Context(), params.logger, owner, repo)
	if err != nil {
		params.presenter.Error("Failed to initialize provider for %s: %v", targetRepo, err)

		return err
	}

	// 5. Construct and create the work item
	newItem, err := constructWorkItem(cmd.Context(), params, owner, repo)
	if err != nil {
		params.presenter.Error("Failed to construct work item: %v", err)

		return err
	}

	// 6. Submit the feedback
	return submitFeedback(cmd.Context(), params.presenter, provider, newItem, targetRepo)
}

// parseFeedbackArgs determines the repository alias and title from command-line arguments.
func parseFeedbackArgs(args []string, defaultRepo string, repositories map[string]string) (string, string) {
	repoAlias := defaultRepo

	var title string

	if len(args) > 0 {
		// If the first arg is a valid repo alias, use it. Otherwise, assume it's the title.
		if _, ok := repositories[args[0]]; ok {
			repoAlias = args[0]
			if len(args) > 1 {
				title = args[1]
			}
		} else {
			title = args[0]
		}
	}

	return repoAlias, title
}

// gatherFeedbackFromUser prompts the user for a title and body if not already provided.
func gatherFeedbackFromUser(title, body *string) error {
	if *title == "" {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("What is the title of your feedback?").Value(title),
				huh.NewText().Title("Please provide more details (optional)").Value(body),
			),
		)
		if err := form.Run(); err != nil {
			return fmt.Errorf("input form failed: %w", err)
		}
	}

	if strings.TrimSpace(*title) == "" {
		return ErrTitleEmpty
	}

	return nil
}

// newProviderForRepo creates a workitem.Provider for a specific owner/repo string.
func newProviderForRepo(ctx context.Context, logger *slog.Logger, owner, repo string) (workitem.Provider, error) {
	ghClient, err := github.NewClient(ctx, logger, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create github api client for %s/%s: %w", owner, repo, err)
	}

	return wigh.NewWithClient(ghClient, logger, owner, repo), nil
}

// constructWorkItem builds the WorkItem struct for submission.
func constructWorkItem(ctx context.Context, params *feedbackParams, owner, repo string) (workitem.WorkItem, error) {
	ghClient, err := github.NewClient(ctx, params.logger, owner, repo)
	if err != nil {
		return workitem.WorkItem{}, fmt.Errorf("failed to create github client for context: %w", err)
	}

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
	}, nil
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
