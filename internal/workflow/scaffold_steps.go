package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/scaffold"
)

// ScaffoldIDXStep generates the .idx configuration files.
type ScaffoldIDXStep struct {
	Presenter PresenterInterface
	AssumeYes bool
}

// Description returns the step description.
func (s *ScaffoldIDXStep) Description() string {
	return "Scaffold Project IDX Environment (.idx/ and .vscode/)"
}

// PreCheck performs pre-flight checks.
func (s *ScaffoldIDXStep) PreCheck(_ context.Context) error { return nil }

// Execute runs the step logic.
func (s *ScaffoldIDXStep) Execute(_ context.Context) error {
	provider := scaffold.NewProvider()

	// 1. Scaffold .idx directory
	idxDir := ".idx"
	//nolint:mnd // 0750 is standard dir permission.
	if err := os.MkdirAll(idxDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .idx directory: %w", err)
	}

	idxFiles, err := provider.GetFiles("idx")
	if err != nil {
		return fmt.Errorf("failed to load idx templates: %w", err)
	}

	if err := writeScaffoldFiles(s.Presenter, s.AssumeYes, idxDir, idxFiles); err != nil {
		return err
	}

	// 2. Scaffold .vscode directory
	vscodeDir := ".vscode"
	//nolint:mnd // 0750 is standard dir permission.
	if err := os.MkdirAll(vscodeDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .vscode directory: %w", err)
	}

	vscodeFiles, err := provider.GetFiles("vscode")
	if err != nil {
		return fmt.Errorf("failed to load vscode templates: %w", err)
	}

	if err := writeScaffoldFiles(s.Presenter, s.AssumeYes, vscodeDir, vscodeFiles); err != nil {
		return err
	}

	s.Presenter.Newline()
	s.Presenter.Advice("Environment scaffolded. You may need to rebuild your environment for changes to take effect.")
	s.Presenter.Advice("Edit .idx/local.nix to set your GPG_KEY_ID.")

	return nil
}

// writeScaffoldFiles is a helper to write a map of files to a directory with confirmation.
func writeScaffoldFiles(presenter PresenterInterface, assumeYes bool, dir string, files map[string]string) error {
	for filename, content := range files {
		path := filepath.Join(dir, filename)
		shouldWrite := true

		// Check existence
		if _, err := os.Stat(path); err == nil {
			if assumeYes {
				presenter.Info("  ! %s exists. Overwriting (due to --yes).", filename)
			} else {
				confirm, _ := presenter.PromptForConfirmation(fmt.Sprintf("  ? %s exists. Overwrite?", filename))
				if !confirm {
					presenter.Info("  ~ Skipping %s.", filename)

					shouldWrite = false
				}
			}
		}

		if shouldWrite {
			//nolint:mnd // 0600 is standard secure file permission.
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				return fmt.Errorf("failed to write %s: %w", filename, err)
			}

			presenter.Success("  + Wrote %s", filename)
		}
	}

	return nil
}

// ScaffoldFirebaseStep initializes Firebase.
type ScaffoldFirebaseStep struct {
	ExecClient *exec.ExecutorClient
	Presenter  PresenterInterface
}

// Description returns the step description.
func (s *ScaffoldFirebaseStep) Description() string {
	return "Scaffold Firebase Environment"
}

// PreCheck performs pre-flight checks.
func (s *ScaffoldFirebaseStep) PreCheck(_ context.Context) error {
	if !s.ExecClient.CommandExists("firebase") {
		s.Presenter.Error("Firebase CLI not found.")
		s.Presenter.Advice("Please rebuild your environment (dev.nix) to include 'firebase-tools'.")
		//nolint:err113 // Dynamic error is appropriate here.
		return errors.New("firebase-tools missing")
	}

	return nil
}

// Execute runs the step logic.
func (s *ScaffoldFirebaseStep) Execute(ctx context.Context) error {
	// Login Check
	_, _, err := s.ExecClient.CaptureOutput(ctx, ".", "firebase", "projects:list", "--json")
	if err != nil {
		s.Presenter.Warning("You do not appear to be logged in to Firebase.")
		s.Presenter.Step("Running 'firebase login'...")

		err = s.ExecClient.Execute(ctx, ".", "firebase", "login")
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	s.Presenter.Step("Initializing Firebase Project Structure...")
	s.Presenter.Info("This will guide you through creating firebase.json and .firebaserc")

	err = s.ExecClient.Execute(ctx, ".", "firebase", "init")
	if err != nil {
		return fmt.Errorf("firebase init failed: %w", err)
	}

	s.Presenter.Success("Firebase environment scaffolded successfully.")

	return nil
}
