// internal/github/client.go
package github

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

const GHTokenEnvVar = "GITHUB_TOKEN"

// Client wraps the go-github client.
type Client struct {
	*github.Client
	logger *slog.Logger
	owner  string
	repo   string
}

// NewClient creates a new GitHub client, authenticating using the GITHUB_TOKEN environment variable.
// It is scoped to a specific repository owner and name.
func NewClient(ctx context.Context, logger *slog.Logger, owner, repo string) (*Client, error) {
	token := os.Getenv(GHTokenEnvVar)
	if token == "" {
		return nil, fmt.Errorf(
			"github token not found: the '%s' environment variable must be set",
			GHTokenEnvVar,
		)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghClient := github.NewClient(tc)

	return &Client{
		Client: ghClient,
		logger: logger,
		owner:  owner,
		repo:   repo,
	}, nil
}

// GetAuthenticatedUserLogin returns the login name of the user authenticated by the token.
func (c *Client) GetAuthenticatedUserLogin(ctx context.Context) (string, error) {
	user, _, err := c.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("could not get authenticated user from token: %w", err)
	}
	login := user.GetLogin()
	if login == "" {
		return "", fmt.Errorf("could not determine user login from token")
	}
	return login, nil
}

// CreateRepo creates a new repository on GitHub for a specific owner (user or org).
// An empty owner string "" defaults to the authenticated user.
func (c *Client) CreateRepo(
	ctx context.Context,
	owner, name, description string,
	isPrivate bool,
) (*github.Repository, error) {
	logOwner := owner
	if logOwner == "" {
		logOwner = "authenticated user"
	}
	c.logger.InfoContext(ctx, "Attempting to create GitHub repository", "owner", logOwner, "repo_name", name, "private", isPrivate)

	repo := &github.Repository{
		Name:        github.String(name),
		Description: github.String(description),
		Private:     github.Bool(isPrivate),
		AutoInit:    github.Bool(true), // Auto-initialize with a README
	}

	createdRepo, resp, err := c.Repositories.Create(ctx, owner, repo)
	if err != nil {
		if errResp, ok := err.(*github.ErrorResponse); ok {
			if len(errResp.Errors) > 0 && errResp.Errors[0].Field == "name" {
				return nil, fmt.Errorf("repository '%s' already exists for owner '%s'", name, logOwner)
			}
		}
		c.logger.ErrorContext(ctx, "Failed to create GitHub repository", "error", err)
		return nil, fmt.Errorf("github API error: %w", err)
	}

	c.logger.DebugContext(ctx, "GitHub API response", "status", resp.Status)
	c.logger.InfoContext(ctx, "Successfully created GitHub repository", "url", createdRepo.GetHTMLURL())

	return createdRepo, nil
}

// UpdateBranchProtection applies a set of protection rules to the client's configured branch.
func (c *Client) UpdateBranchProtection(ctx context.Context, branch string, request github.ProtectionRequest) error {
	c.logger.InfoContext(ctx, "Applying branch protection rules", "owner", c.owner, "repo", c.repo, "branch", branch)

	_, _, err := c.Repositories.UpdateBranchProtection(ctx, c.owner, c.repo, branch, &request)
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to update branch protection", "error", err)

		if strings.Contains(err.Error(), "404 Not Found") {
			return fmt.Errorf("repository '%s/%s' not found or token lacks permission", c.owner, c.repo)
		}
		return fmt.Errorf("github API error while updating branch protection: %w", err)
	}

	c.logger.InfoContext(ctx, "Successfully applied branch protection rules")
	return nil
}
