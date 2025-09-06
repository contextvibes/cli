// internal/github/client.go
package github

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

const GHTokenEnvVar = "GITHUB_TOKEN"

// Client wraps the go-github client.
type Client struct {
	*github.Client
	logger *slog.Logger
}

// NewClient creates a new GitHub client, authenticating using the GITHUB_TOKEN environment variable.
func NewClient(ctx context.Context, logger *slog.Logger) (*Client, error) {
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

	return &Client{
		Client: github.NewClient(tc),
		logger: logger,
	}, nil
}

// CreateRepo creates a new repository on GitHub.
func (c *Client) CreateRepo(
	ctx context.Context,
	name, description string,
	isPrivate bool,
) (*github.Repository, error) {
	c.logger.InfoContext(ctx, "Attempting to create GitHub repository", "repo_name", name, "private", isPrivate)

	repo := &github.Repository{
		Name:        github.String(name),
		Description: github.String(description),
		Private:     github.Bool(isPrivate),
		AutoInit:    github.Bool(true), // Auto-initialize with a README
	}

	createdRepo, resp, err := c.Repositories.Create(ctx, "", repo)
	if err != nil {
		// Check for specific "name already exists" error
		if errResp, ok := err.(*github.ErrorResponse); ok {
			if len(errResp.Errors) > 0 && errResp.Errors[0].Field == "name" {
				return nil, fmt.Errorf("repository '%s' already exists on GitHub", name)
			}
		}
		c.logger.ErrorContext(ctx, "Failed to create GitHub repository", "error", err)
		return nil, fmt.Errorf("github API error: %w", err)
	}

	c.logger.DebugContext(ctx, "GitHub API response", "status", resp.Status)
	c.logger.InfoContext(ctx, "Successfully created GitHub repository", "url", createdRepo.GetHTMLURL())

	return createdRepo, nil
}
