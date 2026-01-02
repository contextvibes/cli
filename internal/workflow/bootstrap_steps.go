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
	Ref        string // The git reference to install (e.g., "main", a hash, or a branch)
}

// Description returns the step description.
func (s *InstallSelfStep) Description() string {
	return fmt.Sprintf("Install ContextVibes CLI (@%s) to ~/go/bin", s.Ref)
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
	// We install from the specified reference to allow testing branches/hashes.
	target := fmt.Sprintf("github.com/contextvibes/cli/cmd/contextvibes@%s", s.Ref)
	err := s.ExecClient.Execute(ctx, ".", "go", "install", target)
	if err != nil {
		return fmt.Errorf("failed to install contextvibes: %w", err)
	}

	return nil
}
