// Package setupidentity provides the command to bootstrap the secure environment.
package setupidentity

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

const (
	dirPermSecure = 0o700
	filePermRW    = 0o600
	minKeyParts   = 5
)

//go:embed setupidentity.md.tpl
var setupIdentityLongDescription string

// SetupIdentityCmd represents the setup-identity command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var SetupIdentityCmd = &cobra.Command{
	Use:   "setup-identity",
	Short: "Bootstraps the secure environment (GPG, Pass, GitHub).",
	//nolint:lll // Long description.
	Long: `Configures the "Chain of Trust" workflow: GPG Agent, Git signing, Password Store, and GitHub CLI authentication.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Secure Environment Bootstrap")

		// --- Phase 1: Plumbing (Configuration) ---
		presenter.Header("1. Configuring Tools & Shell")

		// 1.1 GPG Agent
		if err := configureGPGAgent(ctx, presenter); err != nil {
			return err
		}

		// 1.2 Git Security
		if err := configureGitSecurity(ctx, presenter); err != nil {
			return err
		}

		// 1.3 Bashrc Integration
		if err := configureBashrc(presenter); err != nil {
			return err
		}

		presenter.Newline()

		// --- Phase 2: Identity (Interactive) ---
		presenter.Header("2. Identity & Secrets")

		// 2.1 Import GPG Key
		keyID, err := importGPGKey(ctx, presenter)
		if err != nil {
			return err
		}

		// 2.2 Trust Key
		if err := trustGPGKey(ctx, presenter, keyID); err != nil {
			return err
		}

		// 2.3 Initialize Pass
		if err := initPass(ctx, presenter, keyID); err != nil {
			return err
		}

		// 2.4 GitHub Auth
		if err := authenticateGitHub(ctx, presenter); err != nil {
			return err
		}

		presenter.Success("Bootstrap Complete! Your environment is secure.")
		presenter.Advice("Run 'source ~/.bashrc' to refresh your shell configuration.")

		return nil
	},
}

func configureGPGAgent(ctx context.Context, p *ui.Presenter) error {
	home, _ := os.UserHomeDir()
	gnupgDir := filepath.Join(home, ".gnupg")

	if err := os.MkdirAll(gnupgDir, dirPermSecure); err != nil {
		return fmt.Errorf("failed to create ~/.gnupg: %w", err)
	}

	// Find pinentry-curses
	out, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "which", "pinentry-curses")
	if err != nil {
		p.Warning("pinentry-curses not found. GPG signing might fail in terminal.")
		// Continue anyway
	}

	pinentryPath := strings.TrimSpace(out)
	if pinentryPath != "" {
		confPath := filepath.Join(gnupgDir, "gpg-agent.conf")

		confContent := fmt.Sprintf("pinentry-program %s\n", pinentryPath)

		err := os.WriteFile(confPath, []byte(confContent), filePermRW)
		if err != nil {
			return fmt.Errorf("failed to write gpg-agent.conf: %w", err)
		}
		// Reload agent
		_ = globals.ExecClient.Execute(ctx, ".", "gpg-connect-agent", "reloadagent", "/bye")

		p.Success("✓ GPG Agent configured with %s", pinentryPath)
	}

	return nil
}

func configureGitSecurity(ctx context.Context, p *ui.Presenter) error {
	keyID := os.Getenv("GPG_KEY_ID")
	if keyID == "" {
		p.Info("GPG_KEY_ID env var not set. Skipping automatic Git signing config.")

		return nil
	}

	cmds := [][]string{
		{"config", "--global", "user.signingkey", keyID},
		{"config", "--global", "commit.gpgsign", "true"},
		{"config", "--global", "gpg.program", "gpg"},
	}

	for _, args := range cmds {
		err := globals.ExecClient.Execute(ctx, ".", "git", args...)
		if err != nil {
			return fmt.Errorf("failed to configure git: %w", err)
		}
	}

	p.Success("✓ Git configured to sign commits with key %s", keyID)

	return nil
}

func configureBashrc(p *ui.Presenter) error {
	home, _ := os.UserHomeDir()
	bashrcPath := filepath.Join(home, ".bashrc")
	marker := "# --- SECURE ENV CONFIG ---"

	content, err := os.ReadFile(bashrcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .bashrc: %w", err)
	}

	if strings.Contains(string(content), marker) {
		p.Info("Shell configuration already present.")

		return nil
	}

	block := `
# --- SECURE ENV CONFIG ---
export GPG_TTY=$(tty)

# Status Check
if ! gpg --list-secret-keys --with-colons 2>/dev/null | grep -q "^sec:"; then
    echo " "
    echo "⚠️  IDENTITY NOT FOUND"
    echo "   Run 'contextvibes factory setup-identity' to bootstrap."
    echo " "
else
    echo "✅ Identity Active"
fi

# Aliases
alias p='pass'
alias g='git'
`
	//nolint:gosec // Writing to user's bashrc is intended.
	f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open .bashrc: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(block); err != nil {
		return fmt.Errorf("failed to write to .bashrc: %w", err)
	}

	p.Success("✓ Shell configuration updated (.bashrc)")

	return nil
}

func importGPGKey(ctx context.Context, p *ui.Presenter) (string, error) {
	// Check if key exists
	out, _, _ := globals.ExecClient.CaptureOutput(ctx, ".", "gpg", "--list-secret-keys", "--with-colons")
	if strings.Contains(out, "sec:") {
		p.Info("Secret key already exists. Skipping import.")

		return extractKeyID(out), nil
	}

	p.Info("👉 Please paste your ASCII-Armored Private GPG Key.")
	p.Info("   (Press Enter, paste key, then press Ctrl+D to finish)")

	// Interactive import using standard input
	if err := globals.ExecClient.Execute(ctx, ".", "gpg", "--import"); err != nil {
		return "", fmt.Errorf("gpg import failed: %w", err)
	}

	// Re-check to get ID
	out, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "gpg", "--list-secret-keys", "--with-colons")
	if err != nil {
		return "", fmt.Errorf("failed to list keys after import: %w", err)
	}

	keyID := extractKeyID(out)
	if keyID == "" {
		return "", errors.New("no secret key found after import")
	}

	return keyID, nil
}

func extractKeyID(gpgOutput string) string {
	lines := strings.SplitSeq(gpgOutput, "\n")
	for line := range lines {
		if strings.HasPrefix(line, "sec:") {
			parts := strings.Split(line, ":")
			if len(parts) >= minKeyParts {
				return parts[4]
			}
		}
	}

	return ""
}

func trustGPGKey(ctx context.Context, p *ui.Presenter, keyID string) error {
	p.Step("Applying 'Ultimate Trust' to key: %s", keyID)

	cmdStr := fmt.Sprintf("echo -e \"5\ny\n\" | gpg --command-fd 0 --edit-key %s trust", keyID)

	err := globals.ExecClient.Execute(ctx, ".", "sh", "-c", cmdStr)
	if err != nil {
		p.Warning("Failed to automate trust setting. You may need to trust the key manually.")
	} else {
		p.Success("✓ Key trusted.")
	}

	return nil
}

func initPass(ctx context.Context, p *ui.Presenter, keyID string) error {
	home, _ := os.UserHomeDir()
	passDir := filepath.Join(home, ".password-store")

	if _, err := os.Stat(passDir); err == nil {
		p.Info("Password store already initialized.")

		return nil
	}

	p.Step("Initializing 'pass' vault...")

	err := globals.ExecClient.Execute(ctx, ".", "pass", "init", keyID)
	if err != nil {
		return fmt.Errorf("pass init failed: %w", err)
	}

	p.Success("✓ Vault initialized.")

	return nil
}

func authenticateGitHub(ctx context.Context, p *ui.Presenter) error {
	var token string

	p.Newline()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub Personal Access Token (Fine-grained)").
				Description("Paste your token here. It will be stored securely in 'pass'.").
				EchoMode(huh.EchoModePassword).
				Value(&token),
		),
	)

	err := form.Run()
	if err != nil {
		return fmt.Errorf("input form failed: %w", err)
	}

	if strings.TrimSpace(token) == "" {
		return errors.New("token cannot be empty")
	}

	p.Step("Storing token in vault...")
	// Pipe token to pass insert
	insertCmd := "echo \"" + token + "\" | pass insert -m -f github/token"

	err = globals.ExecClient.Execute(ctx, ".", "sh", "-c", insertCmd)
	if err != nil {
		return fmt.Errorf("failed to store token in pass: %w", err)
	}

	p.Success("✓ Token stored in vault (github/token).")

	p.Step("Authenticating GitHub CLI...")
	// Pipe token to gh auth login
	loginCmd := "echo \"" + token + "\" | gh auth login --with-token"

	err = globals.ExecClient.Execute(ctx, ".", "sh", "-c", loginCmd)
	if err != nil {
		return fmt.Errorf("gh auth login failed: %w", err)
	}

	_ = globals.ExecClient.Execute(ctx, ".", "gh", "auth", "setup-git")

	p.Success("✓ GitHub CLI authenticated.")

	return nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	// Create a default description if the template file is missing or empty during dev
	desc := cmddocs.CommandDesc{
		Short: "Bootstraps the secure environment.",
		Long:  "Configures GPG, Pass, and GitHub Auth.",
	}

	parsed, err := cmddocs.ParseAndExecute(setupIdentityLongDescription, nil)
	if err == nil {
		desc = parsed
	}

	SetupIdentityCmd.Short = desc.Short
	SetupIdentityCmd.Long = desc.Long
}
