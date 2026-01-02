package workflow

import (
	"context"
	"errors"
	"fmt"

	"github.com/contextvibes/cli/internal/exec"
)

// InstallSelfStep installs the contextvibes CLI binary using go install.
type InstallSelfStep struct {
	ExecClient *exec.ExecutorClient
}

// Description returns the step description.
func (s *InstallSelfStep) Description() string {
	return "Install ContextVibes CLI to ~/go/bin"
}

// PreCheck performs pre-flight checks.
func (s *InstallSelfStep) PreCheck(_ context.Context) error {
	if !s.ExecClient.CommandExists("go") {
		//nolint:err113 // Dynamic error is appropriate for CLI prerequisite check.
		return errors.New("go executable not found in PATH; cannot install CLI")
	}

	return nil
}

// Execute runs the installation command.
func (s *InstallSelfStep) Execute(ctx context.Context) error {
	// We install the latest version from the repository.
	// This ensures the user gets the most stable compiled version.
	err := s.ExecClient.Execute(ctx, ".", "go", "install", "github.com/contextvibes/cli/cmd/contextvibes@latest")
	if err != nil {
		return fmt.Errorf("failed to install contextvibes: %w", err)
	}

	return nil
}
