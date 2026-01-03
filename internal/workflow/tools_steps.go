package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
)

// CheckGoEnvStep verifies Go is installed and logs the version.
type CheckGoEnvStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  PresenterInterface
}

// Description returns the step description.
func (s *CheckGoEnvStep) Description() string {
	return "Verify System Go Version"
}

// PreCheck checks if go is in PATH.
func (s *CheckGoEnvStep) PreCheck(_ context.Context) error {
	if !s.ExecClient.CommandExists("go") {
		//nolint:err113 // Dynamic error is appropriate here.
		return errors.New("go executable not found in PATH")
	}

	return nil
}

// Execute runs the step logic.
func (s *CheckGoEnvStep) Execute(ctx context.Context) error {
	out, _, err := s.ExecClient.CaptureOutput(ctx, ".", "go", "version")
	if err != nil {
		return fmt.Errorf("failed to get go version: %w", err)
	}

	s.Presenter.Info("Detected: %s", strings.TrimSpace(out))

	return nil
}

// ConfigurePathStep ensures $HOME/go/bin is in .bashrc.
type ConfigurePathStep struct {
	Presenter PresenterInterface
	AssumeYes bool
}

// Description returns the step description.
func (s *ConfigurePathStep) Description() string {
	return "Configure shell PATH for local Go tools"
}

// PreCheck performs pre-flight checks.
func (s *ConfigurePathStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step logic.
func (s *ConfigurePathStep) Execute(_ context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	rcFile := filepath.Join(home, ".bashrc")

	contentBytes, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", rcFile, err)
	}

	content := string(contentBytes)

	targetLine := "export PATH=$HOME/go/bin:$PATH"
	if strings.Contains(content, targetLine) {
		s.Presenter.Info("PATH configuration already present in %s.", rcFile)

		return nil
	}

	s.Presenter.Warning("Local Go bin path is missing from %s.", rcFile)
	s.Presenter.Info("Proposed addition:")
	s.Presenter.Detail("# Go Tools (Local overrides System/Nix)")
	s.Presenter.Detail(targetLine)

	if !s.AssumeYes {
		confirm, err := s.Presenter.PromptForConfirmation("Append this to your .bashrc?")
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		if !confirm {
			s.Presenter.Info("Skipping PATH configuration.")

			return nil
		}
	}

	//nolint:mnd,gosec // 0644 is standard for .bashrc.
	bashrcFile, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", rcFile, err)
	}

	defer bashrcFile.Close()

	if _, err := bashrcFile.WriteString("\n# Go Tools (Local overrides System/Nix)\n" + targetLine + "\n"); err != nil {
		return fmt.Errorf("failed to write to %s: %w", rcFile, err)
	}

	s.Presenter.Success("Updated %s. Run 'source %s' after this command completes.", rcFile, rcFile)

	return nil
}

// InstallGoToolsStep force-installs the required tools.
type InstallGoToolsStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  PresenterInterface
}

// Description returns the step description.
func (s *InstallGoToolsStep) Description() string {
	return "Force rebuild and install Go tools"
}

// PreCheck performs pre-flight checks.
func (s *InstallGoToolsStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step logic.
func (s *InstallGoToolsStep) Execute(ctx context.Context) error {
	tools := []string{
		"golang.org/x/vuln/cmd/govulncheck@latest",
		"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest",
		"golang.org/x/tools/cmd/deadcode@latest",
		"golang.org/x/tools/cmd/goimports@latest",
		"golang.org/x/tools/cmd/stringer@latest",
		"golang.org/x/tools/cmd/godoc@latest",
	}

	// Ensure GOBIN is set for this session so installs go to the right place
	home, _ := os.UserHomeDir()
	goBin := filepath.Join(home, "go", "bin")

	for _, tool := range tools {
		s.Presenter.Step("Installing %s...", tool)
		// -a forces rebuild
		err := s.ExecClient.Execute(ctx, ".", "go", "install", "-a", tool)
		if err != nil {
			s.Presenter.Error("Failed to install %s: %v", tool, err)

			return fmt.Errorf("failed to install %s: %w", tool, err)
		}
	}

	// Verification
	s.Presenter.Newline()
	s.Presenter.Info("Verifying govulncheck resolution...")

	// We check where the command resolves *now*
	out, _, _ := s.ExecClient.CaptureOutput(ctx, ".", "which", "govulncheck")
	resolvedPath := strings.TrimSpace(out)
	expectedPath := filepath.Join(goBin, "govulncheck")

	if resolvedPath == expectedPath {
		s.Presenter.Success("govulncheck resolves to %s", resolvedPath)
	} else {
		s.Presenter.Warning("govulncheck resolves to %s", resolvedPath)
		s.Presenter.Advice("Expected: %s", expectedPath)
		s.Presenter.Advice("Please run 'source ~/.bashrc' to update your current shell.")
	}

	return nil
}
