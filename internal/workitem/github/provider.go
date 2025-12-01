// internal/workitem/github/provider.go
package github

import (
	"context"
	_ "errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/git"
	gh "github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/google/go-github/v74/github"
)

// Provider implements the workitem.Provider interface for GitHub Issues.
type Provider struct {
	ghClient *gh.Client
	logger   *slog.Logger
	owner    string
	repo     string
}

// NewWithClient creates a new Provider with an existing GitHub client.
func NewWithClient(client *gh.Client, logger *slog.Logger, owner, repo string) workitem.Provider {
	return &Provider{
		ghClient: client,
		logger:   logger,
		owner:    owner,
		repo:     repo,
	}
}

// New creates a new Provider by discovering the repository from the local git remote.
func New(ctx context.Context, logger *slog.Logger, cfg *config.Config) (workitem.Provider, error) {
	tempExecutor := exec.NewOSCommandExecutor(logger)

	gitClient, err := git.NewClient(
		ctx,
		".",
		git.GitClientConfig{Executor: tempExecutor, Logger: logger},
	)
	if err != nil {
		return nil, fmt.Errorf("could not initialize git client for repo discovery: %w", err)
	}

	remoteURL, err := gitClient.GetRemoteURL(ctx, cfg.Git.DefaultRemote)
	if err != nil {
		return nil, fmt.Errorf("could not get remote URL for '%s': %w", cfg.Git.DefaultRemote, err)
	}

	owner, repo, err := gh.ParseGitHubRemote(remoteURL)
	if err != nil {
		return nil, fmt.Errorf(
			"could not parse owner/repo from remote URL '%s': %w",
			remoteURL,
			err,
		)
	}

	logger.DebugContext(
		ctx,
		"Discovered GitHub repository from remote",
		"owner",
		owner,
		"repo",
		repo,
	)

	ghClient, err := gh.NewClient(ctx, logger, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create github api client: %w", err)
	}

	return NewWithClient(ghClient, logger, owner, repo), nil
}

func (p *Provider) ListItems(
	ctx context.Context,
	options workitem.ListOptions,
) ([]workitem.WorkItem, error) {
	ghOpts := &github.IssueListByRepoOptions{
		State:    "open",
		Labels:   options.Labels,
		Assignee: options.Assignee,
		ListOptions: github.ListOptions{
			PerPage: options.Limit,
			Page:    options.Page,
		},
	}
	if options.State == workitem.StateClosed {
		ghOpts.State = "closed"
	}

	p.logger.DebugContext(
		ctx,
		"Listing GitHub issues",
		"owner",
		p.owner,
		"repo",
		p.repo,
		"options",
		ghOpts,
	)

	issues, _, err := p.ghClient.Issues.ListByRepo(ctx, p.owner, p.repo, ghOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list github issues: %w", err)
	}

	workItems := make([]workitem.WorkItem, 0, len(issues))
	for _, issue := range issues {
		if issue.IsPullRequest() {
			continue
		}

		workItems = append(workItems, toWorkItem(issue))
	}

	return workItems, nil
}

func (p *Provider) GetItem(
	ctx context.Context,
	number int,
	withComments bool,
) (*workitem.WorkItem, error) {
	p.logger.DebugContext(
		ctx,
		"Getting GitHub issue",
		"owner",
		p.owner,
		"repo",
		p.repo,
		"number",
		number,
	)

	issue, _, err := p.ghClient.Issues.Get(ctx, p.owner, p.repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get github issue #%d: %w", number, err)
	}

	item := toWorkItem(issue)
	if withComments && issue.GetComments() > 0 {
		p.logger.DebugContext(ctx, "Fetching comments for issue", "number", number)

		comments, _, err := p.ghClient.Issues.ListComments(ctx, p.owner, p.repo, number, nil)
		if err != nil {
			p.logger.WarnContext(
				ctx,
				"Failed to fetch comments for issue",
				"number",
				number,
				"error",
				err,
			)
		} else {
			item.Comments = make([]workitem.Comment, 0, len(comments))
			for _, comment := range comments {
				item.Comments = append(item.Comments, toComment(comment))
			}
		}
	}

	return &item, nil
}

func (p *Provider) CreateItem(
	ctx context.Context,
	item workitem.WorkItem,
) (*workitem.WorkItem, error) {
	issueReq := fromWorkItem(item)
	p.logger.DebugContext(
		ctx,
		"Creating GitHub issue",
		"owner",
		p.owner,
		"repo",
		p.repo,
		"title",
		issueReq.GetTitle(),
	)

	createdIssue, _, err := p.ghClient.Issues.Create(ctx, p.owner, p.repo, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create github issue: %w", err)
	}

	newItem := toWorkItem(createdIssue)

	return &newItem, nil
}

func (p *Provider) UpdateItem(
	ctx context.Context,
	number int,
	item workitem.WorkItem,
) (*workitem.WorkItem, error) {
	issueReq := fromWorkItem(item)

	p.logger.DebugContext(
		ctx,
		"Updating GitHub issue",
		"owner",
		p.owner,
		"repo",
		p.repo,
		"number",
		number,
	)

	updatedIssue, _, err := p.ghClient.Issues.Edit(ctx, p.owner, p.repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update github issue #%d: %w", number, err)
	}

	newItem := toWorkItem(updatedIssue)

	return &newItem, nil
}

func (p *Provider) SearchItems(ctx context.Context, query string) ([]workitem.WorkItem, error) {
	fullQuery := fmt.Sprintf("repo:%s/%s %s", p.owner, p.repo, query)
	p.logger.DebugContext(ctx, "Searching GitHub issues", "query", fullQuery)

	result, _, err := p.ghClient.Search.Issues(ctx, fullQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search github issues: %w", err)
	}

	workItems := make([]workitem.WorkItem, 0, len(result.Issues))
	for _, issue := range result.Issues {
		if issue.IsPullRequest() {
			continue
		}

		workItems = append(workItems, toWorkItem(issue))
	}

	return workItems, nil
}

func (p *Provider) CreateLabel(ctx context.Context, label workitem.Label) (*workitem.Label, error) {
	p.logger.DebugContext(
		ctx,
		"Creating GitHub label",
		"owner",
		p.owner,
		"repo",
		p.repo,
		"name",
		label.Name,
	)
	ghLabel := &github.Label{
		Name:        github.Ptr(label.Name),
		Description: github.Ptr(label.Description),
		Color:       github.Ptr(label.Color),
	}

	createdLabel, _, err := p.ghClient.Issues.CreateLabel(ctx, p.owner, p.repo, ghLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to create github label: %w", err)
	}

	return &workitem.Label{
		Name:        createdLabel.GetName(),
		Description: createdLabel.GetDescription(),
		Color:       createdLabel.GetColor(),
	}, nil
}

func toComment(comment *github.IssueComment) workitem.Comment {
	return workitem.Comment{
		Author:    comment.GetUser().GetLogin(),
		Body:      comment.GetBody(),
		CreatedAt: comment.GetCreatedAt().Time,
		URL:       comment.GetHTMLURL(),
	}
}

func toWorkItem(issue *github.Issue) workitem.WorkItem {
	item := workitem.WorkItem{
		ID:        issue.GetNodeID(),
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		URL:       issue.GetHTMLURL(),
		Author:    issue.GetUser().GetLogin(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
	}
	if issue.GetState() == "closed" {
		item.State = workitem.StateClosed
	} else {
		item.State = workitem.StateOpen
	}

	for _, label := range issue.Labels {
		labelName := label.GetName()

		item.Labels = append(item.Labels, labelName)
		switch strings.ToLower(labelName) {
		case "epic":
			item.Type = workitem.TypeEpic
		case "story", "user story":
			item.Type = workitem.TypeStory
		case "bug":
			item.Type = workitem.TypeBug
		case "chore":
			item.Type = workitem.TypeChore
		}
	}

	if item.Type == "" {
		item.Type = workitem.TypeTask
	}

	for _, assignee := range issue.Assignees {
		item.Assignees = append(item.Assignees, assignee.GetLogin())
	}

	return item
}

// fromWorkItem is the corrected function that prevents sending 'null' for empty slices.
func fromWorkItem(item workitem.WorkItem) *github.IssueRequest {
	// THE FIX: Ensure slices are non-nil for JSON marshaling.
	// If the source slice is nil, we replace it with an empty, non-nil slice.
	labels := item.Labels
	if labels == nil {
		labels = []string{}
	}

	assignees := item.Assignees
	if assignees == nil {
		assignees = []string{}
	}

	req := &github.IssueRequest{
		Title:     github.Ptr(item.Title),
		Body:      github.Ptr(item.Body),
		Labels:    &labels,
		Assignees: &assignees,
	}

	// This part of the logic remains the same.
	typeLabel := strings.ToLower(string(item.Type))
	hasTypeLabel := false

	for _, l := range item.Labels {
		if strings.ToLower(l) == typeLabel {
			hasTypeLabel = true

			break
		}
	}

	if !hasTypeLabel && typeLabel != "" && item.Type != workitem.TypeTask {
		*req.Labels = append(*req.Labels, typeLabel)
	}

	return req
}
