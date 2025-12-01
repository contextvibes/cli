// Package bootstrap provides the steps for bootstrapping a new project.
package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/exec"
	gh "github.com/contextvibes/cli/internal/github" // aliased to avoid conflict
	"github.com/contextvibes/cli/internal/workflow"
)

// --- Step 1: Create Remote Repository ---

// CreateRemoteRepoStep creates a new remote repository on GitHub.
type CreateRemoteRepoStep struct {
	GHClient        *gh.Client
	Presenter       workflow.PresenterInterface
	Owner           string // ADDED Owner field
	RepoName        string
	RepoDescription string
	IsPrivate       bool
	// Outputs of this step to be used by subsequent steps
	CreatedRepoURL string
	CloneURL       string
}

// Description returns a description of the step.
func (s *CreateRemoteRepoStep) Description() string {
	return fmt.Sprintf("Create remote GitHub repository: %s/%s", s.Owner, s.RepoName)
}

// PreCheck performs pre-execution checks.
func (s *CreateRemoteRepoStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *CreateRemoteRepoStep) Execute(ctx context.Context) error {
	// CORRECTED: Pass the owner to the CreateRepo method
	repo, err := s.GHClient.CreateRepo(ctx, s.Owner, s.RepoName, s.RepoDescription, s.IsPrivate)
	if err != nil {
		s.Presenter.Error("Failed to create GitHub repository: %v", err)

		return fmt.Errorf("failed to create repo: %w", err)
	}

	s.CreatedRepoURL = repo.GetHTMLURL()
	s.CloneURL = repo.GetCloneURL()
	s.Presenter.Success("✓ Remote repository created at %s", s.CreatedRepoURL)

	return nil
}

// --- Step 2: Clone Repository ---

// CloneRepoStep clones a remote repository to a local path.
type CloneRepoStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  workflow.PresenterInterface
	CloneURL   string
	LocalPath  string
}

// Description returns a description of the step.
func (s *CloneRepoStep) Description() string {
	return "Clone repository to local path: ./" + s.LocalPath
}

// PreCheck performs pre-execution checks.
func (s *CloneRepoStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *CloneRepoStep) Execute(ctx context.Context) error {
	err := s.ExecClient.Execute(ctx, ".", "git", "clone", s.CloneURL, s.LocalPath)
	if err != nil {
		s.Presenter.Error("Failed to clone repository: %v", err)

		return fmt.Errorf("failed to clone repo: %w", err)
	}

	return nil
}

// --- Step 3: Scaffold Project ---

// ScaffoldProjectStep scaffolds the project structure and template files.
type ScaffoldProjectStep struct {
	Presenter    workflow.PresenterInterface
	LocalPath    string
	GoModulePath string
	AppName      string
}

// Description returns a description of the step.
func (s *ScaffoldProjectStep) Description() string {
	return "Scaffold project structure and template files"
}

// PreCheck performs pre-execution checks.
func (s *ScaffoldProjectStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *ScaffoldProjectStep) Execute(_ context.Context) error {
	readmePath := filepath.Join(s.LocalPath, "README.md")
	content := fmt.Sprintf("# %s\n\nGo Module: `%s`\n", s.AppName, s.GoModulePath)

	//nolint:gosec // Writing README with 0644 is standard.
	err := os.WriteFile(readmePath, []byte(content), 0o644)
	if err != nil {
		s.Presenter.Error("Failed to write placeholder README.md: %v", err)

		return fmt.Errorf("failed to write README: %w", err)
	}

	s.Presenter.Success("✓ Placeholder README.md created.")

	return nil
}

// --- Step 4 & 5: Initial Commit and Push ---

// InitialCommitAndPushStep creates and pushes the initial commit.
type InitialCommitAndPushStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  workflow.PresenterInterface
	LocalPath  string
}

// Description returns a description of the step.
func (s *InitialCommitAndPushStep) Description() string { return "Create and push initial commit" }

// PreCheck performs pre-execution checks.
func (s *InitialCommitAndPushStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step.
func (s *InitialCommitAndPushStep) Execute(ctx context.Context) error {
	err := s.ExecClient.Execute(ctx, s.LocalPath, "git", "add", ".")
	if err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	err = s.ExecClient.Execute(ctx, s.LocalPath, "git", "commit", "-m", "Initial commit")
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	err = s.ExecClient.Execute(ctx, s.LocalPath, "git", "push")
	if err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}
