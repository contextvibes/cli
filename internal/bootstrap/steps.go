// internal/bootstrap/steps.go
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

func (s *CreateRemoteRepoStep) Description() string {
	return fmt.Sprintf("Create remote GitHub repository: %s/%s", s.Owner, s.RepoName)
}

func (s *CreateRemoteRepoStep) PreCheck(ctx context.Context) error { return nil }

func (s *CreateRemoteRepoStep) Execute(ctx context.Context) error {
	// CORRECTED: Pass the owner to the CreateRepo method
	repo, err := s.GHClient.CreateRepo(ctx, s.Owner, s.RepoName, s.RepoDescription, s.IsPrivate)
	if err != nil {
		s.Presenter.Error("Failed to create GitHub repository: %v", err)
		return err
	}
	s.CreatedRepoURL = repo.GetHTMLURL()
	s.CloneURL = repo.GetCloneURL()
	s.Presenter.Success("✓ Remote repository created at %s", s.CreatedRepoURL)
	return nil
}

// --- Step 2: Clone Repository ---

type CloneRepoStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  workflow.PresenterInterface
	CloneURL   string
	LocalPath  string
}

func (s *CloneRepoStep) Description() string {
	return fmt.Sprintf("Clone repository to local path: ./%s", s.LocalPath)
}
func (s *CloneRepoStep) PreCheck(ctx context.Context) error { return nil }
func (s *CloneRepoStep) Execute(ctx context.Context) error {
	err := s.ExecClient.Execute(ctx, ".", "git", "clone", s.CloneURL, s.LocalPath)
	if err != nil {
		s.Presenter.Error("Failed to clone repository: %v", err)
		return err
	}
	return nil
}

// --- Step 3: Scaffold Project ---

type ScaffoldProjectStep struct {
	Presenter    workflow.PresenterInterface
	LocalPath    string
	GoModulePath string
	AppName      string
}

func (s *ScaffoldProjectStep) Description() string {
	return "Scaffold project structure and template files"
}
func (s *ScaffoldProjectStep) PreCheck(ctx context.Context) error { return nil }
func (s *ScaffoldProjectStep) Execute(ctx context.Context) error {
	readmePath := filepath.Join(s.LocalPath, "README.md")
	content := fmt.Sprintf("# %s\n\nGo Module: `%s`\n", s.AppName, s.GoModulePath)

	err := os.WriteFile(readmePath, []byte(content), 0o644)
	if err != nil {
		s.Presenter.Error("Failed to write placeholder README.md: %v", err)
		return err
	}
	s.Presenter.Success("✓ Placeholder README.md created.")
	return nil
}

// --- Step 4 & 5: Initial Commit and Push ---

type InitialCommitAndPushStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  workflow.PresenterInterface
	LocalPath  string
}

func (s *InitialCommitAndPushStep) Description() string                { return "Create and push initial commit" }
func (s *InitialCommitAndPushStep) PreCheck(ctx context.Context) error { return nil }
func (s *InitialCommitAndPushStep) Execute(ctx context.Context) error {
	if err := s.ExecClient.Execute(ctx, s.LocalPath, "git", "add", "."); err != nil {
		return err
	}
	if err := s.ExecClient.Execute(ctx, s.LocalPath, "git", "commit", "-m", "Initial commit"); err != nil {
		return err
	}
	if err := s.ExecClient.Execute(ctx, s.LocalPath, "git", "push"); err != nil {
		return err
	}
	return nil
}
