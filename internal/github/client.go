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

	"github.com/contextvibes/cli/internal/exec"
	"github.com/google/go-github/v74/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GHTokenEnvVar is the environment variable name for the GitHub token.
//
//nolint:gosec // This is a variable name, not a credential.
const GHTokenEnvVar = "GITHUB_TOKEN"

// PassTokenKey is the key used in the password store.
const PassTokenKey = "github/token"

var (
	sshRemoteRegex = regexp.MustCompile(`^git@github\.com:([\w-]+)/([\w-]+)\.git$`)

	// ErrInvalidRemoteURL is returned when the remote URL is invalid.
	ErrInvalidRemoteURL = errors.New("invalid remote URL")
	// ErrTokenNotFound is returned when the GitHub token is not found.
	ErrTokenNotFound = errors.New("GitHub token not found")
	// ErrProjectNotFound is returned when a project is not found.
	ErrProjectNotFound = errors.New("project not found")
	// ErrInvalidProjectID is returned when a project ID is invalid.
	ErrInvalidProjectID = errors.New("invalid project ID")
	// ErrUserLoginNotFound is returned when the user login cannot be determined.
	ErrUserLoginNotFound = errors.New("could not determine user login")
	// ErrRepoAlreadyExists is returned when a repository already exists.
	ErrRepoAlreadyExists = errors.New("repository already exists")
	// ErrRepoNotFoundOrAuth is returned when a repository is not found or auth fails.
	ErrRepoNotFoundOrAuth = errors.New("repository not found or token lacks permission")
	// ErrPassCommandNotFound is returned when the 'pass' command is not available.
	ErrPassCommandNotFound = errors.New("'pass' command not found")
	// ErrPassOutputEmpty is returned when 'pass' returns no output.
	ErrPassOutputEmpty = errors.New("pass output was empty")
)

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
//
//nolint:nonamedreturns // Named returns are used for clarity in return signature.
func ParseGitHubRemote(remoteURL string) (owner, repo string, err error) {
	//nolint:mnd // Regex match count 3 is specific to this pattern.
	if matches := sshRemoteRegex.FindStringSubmatch(remoteURL); len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	parsed, err := url.Parse(remoteURL)
	if err != nil {
		return "", "", fmt.Errorf("could not parse remote URL: %w", err)
	}

	if parsed.Hostname() != "github.com" {
		return "", "", fmt.Errorf("%w: not a github.com URL: %s", ErrInvalidRemoteURL, parsed.Hostname())
	}

	pathParts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	//nolint:mnd // Expecting at least owner and repo.
	if len(pathParts) < 2 {
		return "", "", fmt.Errorf("%w: path does not contain owner/repo: %s", ErrInvalidRemoteURL, parsed.Path)
	}

	repo = strings.TrimSuffix(pathParts[1], ".git")

	return pathParts[0], repo, nil
}

// NewClient creates a new GitHub client.
// It first checks the GITHUB_TOKEN environment variable.
// If not found, it attempts to retrieve the token from 'pass' (github/token).
func NewClient(ctx context.Context, logger *slog.Logger, owner, repo string) (*Client, error) {
	token := os.Getenv(GHTokenEnvVar)

	if token == "" {
		logger.DebugContext(ctx, "GITHUB_TOKEN env var empty, attempting to retrieve from 'pass'...")

		passToken, err := tryFetchTokenFromPass(ctx, logger)
		if err == nil && passToken != "" {
			token = passToken

			logger.InfoContext(ctx, "Successfully retrieved GitHub token from 'pass'")
		} else if err != nil {
			logger.DebugContext(ctx, "Failed to retrieve from pass", "error", err)
		}
	}

	if token == "" {
		errorMsg := `
GitHub token not found. 

1. Ensure you have run 'contextvibes factory setup-identity' to store your token in 'pass'.
2. OR export it manually: export GITHUB_TOKEN="your_token_here"`

		return nil, fmt.Errorf("%w: %s", ErrTokenNotFound, strings.TrimSpace(errorMsg))
	}

	ts := oauth2.StaticTokenSource(
		//nolint:exhaustruct // Only AccessToken is needed.
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

func tryFetchTokenFromPass(ctx context.Context, logger *slog.Logger) (string, error) {
	// Create a temporary executor just for this check
	executor := exec.NewOSCommandExecutor(logger)
	client := exec.NewClient(executor)

	if !client.CommandExists("pass") {
		return "", ErrPassCommandNotFound
	}

	stdout, _, err := client.CaptureOutput(ctx, ".", "pass", "show", PassTokenKey)
	if err != nil {
		return "", fmt.Errorf("failed to capture pass output: %w", err)
	}

	// pass output might contain multiple lines, the first line is the secret
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "", ErrPassOutputEmpty
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
		} `graphql:"organization(login: )"`
	}

	variables := map[string]any{"owner": githubv4.String(c.owner)}

	c.logger.DebugContext(ctx, "Executing GraphQL query for projects", "owner", c.owner)

	err := c.GraphQL.Query(ctx, &query, variables)
	if err != nil {
		c.logger.ErrorContext(ctx, "GraphQL query for projects failed", "error", err)

		return nil, fmt.Errorf("failed to query for projects: %w", err)
	}

	//nolint:prealloc // Pre-allocating is good but not strictly required for small lists.
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
			} `graphql:"projectV2(number: )"`
		} `graphql:"organization(login: )"`
	}

	variables := map[string]any{
		"owner": githubv4.String(c.owner),
		//nolint:gosec // G115: Project number is unlikely to overflow int32.
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
		return nil, fmt.Errorf("%w: #%d for owner '%s'", ErrProjectNotFound, number, c.owner)
	}

	//nolint:varnamelen // 'id' is standard for identifier.
	id, ok := project.ID.(string)
	if !ok {
		return nil, fmt.Errorf("%w: #%d is not a string", ErrInvalidProjectID, number)
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
		//nolint:revive // GraphQL field name must match schema.
		AddProjectV2ItemById struct {
			Item struct {
				ID githubv4.ID
			}
		} `graphql:"addProjectV2ItemById(input: {projectId: , contentId: })"`
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
		return "", ErrUserLoginNotFound
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

	//nolint:exhaustruct // Partial initialization is standard for GitHub API requests.
	repo := &github.Repository{
		Name:        github.Ptr(name),
		Description: github.Ptr(description),
		Private:     github.Ptr(isPrivate),
		AutoInit:    github.Ptr(true),
	}

	createdRepo, resp, err := c.Repositories.Create(ctx, owner, repo)
	if err != nil {
		//nolint:exhaustruct // Partial initialization for error checking.
		errResp := &github.ErrorResponse{}
		if errors.As(err, &errResp) {
			if len(errResp.Errors) > 0 && errResp.Errors[0].Field == "name" {
				return nil, fmt.Errorf(
					"%w: '%s' for owner '%s'",
					ErrRepoAlreadyExists,
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
				"%w: '%s/%s'",
				ErrRepoNotFoundOrAuth,
				c.owner,
				c.repo,
			)
		}

		return fmt.Errorf("github API error while updating branch protection: %w", err)
	}

	c.logger.InfoContext(ctx, "Successfully applied branch protection rules")

	return nil
}
