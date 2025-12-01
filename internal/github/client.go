// internal/github/client.go
package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v74/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// gosec:G101
const GHTokenEnvVar = "GITHUB_TOKEN"

var sshRemoteRegex = regexp.MustCompile(`^git@github\.com:([\w-]+)/([\w-]+)\.git$`)

// Client wraps the go-github clients for both REST and GraphQL APIs.
type Client struct {
	*github.Client

	GraphQL *githubv4.Client
	logger  *slog.Logger
	owner   string
	repo    string
}

// Project represents a GitHub Project (V2) for listing.
type Project struct {
	Title  string
	Number int
	URL    string
}

// ProjectWithID represents a GitHub Project (V2) with its GraphQL Node ID.
type ProjectWithID struct {
	ID     string
	Title  string
	Number int
	URL    string
}

// ParseGitHubRemote extracts the owner and repository name from a GitHub remote URL.
func ParseGitHubRemote(remoteURL string) (owner, repo string, err error) {
	if matches := sshRemoteRegex.FindStringSubmatch(remoteURL); len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	parsed, err := url.Parse(remoteURL)
	if err != nil {
		return "", "", fmt.Errorf("could not parse remote URL: %w", err)
	}

	if parsed.Hostname() != "github.com" {
		return "", "", fmt.Errorf("remote URL is not a github.com URL: %s", parsed.Hostname())
	}

	pathParts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(pathParts) < 2 {
		return "", "", fmt.Errorf("remote URL path does not contain owner/repo: %s", parsed.Path)
	}

	repo = strings.TrimSuffix(pathParts[1], ".git")

	return pathParts[0], repo, nil
}

// NewClient creates a new GitHub client, authenticating using the GITHUB_TOKEN environment variable.
func NewClient(ctx context.Context, logger *slog.Logger, owner, repo string) (*Client, error) {
	token := os.Getenv(GHTokenEnvVar)
	if token == "" {
		errorMsg := `
GitHub token not found. Please create a Personal Access Token (classic) with 'repo' and 'project' scopes.
See: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens

Then, export it as an environment variable:
export GITHUB_TOKEN="your_token_here"`

		return nil, errors.New(strings.TrimSpace(errorMsg))
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(httpClient)
	ghGraphQLClient := githubv4.NewClient(httpClient)

	return &Client{
		Client:  ghClient,
		GraphQL: ghGraphQLClient,
		logger:  logger,
		owner:   owner,
		repo:    repo,
	}, nil
}

// ListProjects fetches the first 100 GitHub Projects (V2) for the repository's owner.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var query struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes []struct {
					Title  githubv4.String
					Number githubv4.Int
					URL    githubv4.URI
				}
			} `graphql:"projectsV2(first: 100)"`
		} `graphql:"organization(login: $owner)"`
	}

	variables := map[string]any{"owner": githubv4.String(c.owner)}

	c.logger.DebugContext(ctx, "Executing GraphQL query for projects", "owner", c.owner)

	err := c.GraphQL.Query(ctx, &query, variables)
	if err != nil {
		c.logger.ErrorContext(ctx, "GraphQL query for projects failed", "error", err)

		return nil, fmt.Errorf("failed to query for projects: %w", err)
	}

	var projects []Project
	for _, p := range query.Organization.ProjectsV2.Nodes {
		projects = append(projects, Project{
			Title:  string(p.Title),
			Number: int(p.Number),
			URL:    p.URL.String(),
		})
	}

	return projects, nil
}

// GetProjectByNumber fetches a single project by its number to get its GraphQL Node ID.
func (c *Client) GetProjectByNumber(ctx context.Context, number int) (*ProjectWithID, error) {
	var query struct {
		Organization struct {
			ProjectV2 struct {
				ID     githubv4.ID
				Title  githubv4.String
				Number githubv4.Int
				URL    githubv4.URI
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}

	variables := map[string]any{
		"owner":  githubv4.String(c.owner),
		"number": githubv4.Int(number),
	}

	c.logger.DebugContext(
		ctx,
		"Executing GraphQL query for single project",
		"owner",
		c.owner,
		"number",
		number,
	)

	err := c.GraphQL.Query(ctx, &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query for project #%d: %w", number, err)
	}

	project := query.Organization.ProjectV2
	if project.ID == nil {
		return nil, fmt.Errorf("project #%d not found for owner '%s'", number, c.owner)
	}

	id, ok := project.ID.(string)
	if !ok {
		return nil, fmt.Errorf("project ID for #%d is not a string", number)
	}

	return &ProjectWithID{
		ID:     id,
		Title:  string(project.Title),
		Number: int(project.Number),
		URL:    project.URL.String(),
	}, nil
}

// AddIssueToProject adds an issue (by its GraphQL Node ID) to a project (by its GraphQL Node ID).
func (c *Client) AddIssueToProject(ctx context.Context, projectID string, issueID string) error {
	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID githubv4.ID
			}
		} `graphql:"addProjectV2ItemById(input: {projectId: $projectID, contentId: $issueID})"`
	}

	variables := map[string]any{
		"projectID": githubv4.ID(projectID),
		"issueID":   githubv4.ID(issueID),
	}

	c.logger.DebugContext(
		ctx,
		"Executing GraphQL mutation to add issue to project",
		"projectID",
		projectID,
		"issueID",
		issueID,
	)

	err := c.GraphQL.Mutate(ctx, &mutation, variables, nil)
	if err != nil {
		return fmt.Errorf("failed to add issue %s to project %s: %w", issueID, projectID, err)
	}

	return nil
}

// GetAuthenticatedUserLogin returns the login name of the user authenticated by the token.
func (c *Client) GetAuthenticatedUserLogin(ctx context.Context) (string, error) {
	user, _, err := c.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("could not get authenticated user from token: %w", err)
	}

	login := user.GetLogin()
	if login == "" {
		return "", errors.New("could not determine user login from token")
	}

	return login, nil
}

// CreateRepo creates a new repository on GitHub for a specific owner (user or org).
func (c *Client) CreateRepo(
	ctx context.Context,
	owner, name, description string,
	isPrivate bool,
) (*github.Repository, error) {
	logOwner := owner
	if logOwner == "" {
		logOwner = "authenticated user"
	}

	c.logger.InfoContext(
		ctx,
		"Attempting to create GitHub repository",
		"owner",
		logOwner,
		"repo_name",
		name,
		"private",
		isPrivate,
	)

	repo := &github.Repository{
		Name:        github.Ptr(name),
		Description: github.Ptr(description),
		Private:     github.Ptr(isPrivate),
		AutoInit:    github.Ptr(true),
	}

	createdRepo, resp, err := c.Repositories.Create(ctx, owner, repo)
	if err != nil {
		errResp := &github.ErrorResponse{}
		if errors.As(err, &errResp) {
			if len(errResp.Errors) > 0 && errResp.Errors[0].Field == "name" {
				return nil, fmt.Errorf(
					"repository '%s' already exists for owner '%s'",
					name,
					logOwner,
				)
			}
		}

		c.logger.ErrorContext(ctx, "Failed to create GitHub repository", "error", err)

		return nil, fmt.Errorf("github API error: %w", err)
	}

	c.logger.DebugContext(ctx, "GitHub API response", "status", resp.Status)
	c.logger.InfoContext(
		ctx,
		"Successfully created GitHub repository",
		"url",
		createdRepo.GetHTMLURL(),
	)

	return createdRepo, nil
}

// UpdateBranchProtection applies a set of protection rules to the client's configured branch.
func (c *Client) UpdateBranchProtection(
	ctx context.Context,
	branch string,
	request github.ProtectionRequest,
) error {
	c.logger.InfoContext(
		ctx,
		"Applying branch protection rules",
		"owner",
		c.owner,
		"repo",
		c.repo,
		"branch",
		branch,
	)

	_, _, err := c.Repositories.UpdateBranchProtection(ctx, c.owner, c.repo, branch, &request)
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to update branch protection", "error", err)

		if strings.Contains(err.Error(), "404 Not Found") {
			return fmt.Errorf(
				"repository '%s/%s' not found or token lacks permission",
				c.owner,
				c.repo,
			)
		}

		return fmt.Errorf("github API error while updating branch protection: %w", err)
	}

	c.logger.InfoContext(ctx, "Successfully applied branch protection rules")

	return nil
}
