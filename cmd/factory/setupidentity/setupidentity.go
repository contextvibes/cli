// Package setupidentity provides the command to bootstrap the secure environment.
package setupidentity

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

const (
	dirPermSecure       = 0o700
	filePermRW          = 0o600
	filePermRead        = 0o644
	minKeyParts         = 5
	fingerprintPartIdx  = 9
	fingerprintMinParts = 10
)

var (
	// ErrFingerprintNotFound is returned when a GPG key fingerprint cannot be parsed.
	ErrFingerprintNotFound = errors.New("could not determine fingerprint for key")
	// ErrNoSecretKey is returned when no secret key is found after import.
	ErrNoSecretKey = errors.New("no secret key found after import")
	// ErrEmptyToken is returned when the user provides an empty token.
	ErrEmptyToken = errors.New("token cannot be empty")
)

//go:embed setupidentity.md.tpl
var setupIdentityLongDescription string

// NewSetupIdentityCmd creates and configures the `setup-identity` command.
func NewSetupIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setup-identity",
		Short:   "Bootstraps the secure environment (GPG, Pass, GitHub).",
		Long:    "Configures GPG, Git signing, Password Store, and GitHub CLI.",
		Example: `  contextvibes factory setup-identity`,
		GroupID: "factory",
		RunE:    runSetupIdentity,

		// Boilerplate
		Aliases:                    []string{},
		SuggestFor:                 []string{},
		ValidArgs:                  []string{},
		ValidArgsFunction:          nil,
		Args:                       nil,
		ArgAliases:                 []string{},
		BashCompletionFunction:     "",
		Deprecated:                 "",
		Annotations:                nil,
		Version:                    "",
		PersistentPreRun:           nil,
		PersistentPreRunE:          nil,
		PreRun:                     nil,
		PostRun:                    nil,
		PostRunE:                   nil,
		PersistentPostRun:          nil,
		PersistentPostRunE:         nil,
		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
		CompletionOptions:          cobra.CompletionOptions{},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              true,
		SilenceUsage:               true,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          true,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}

	// Set the long description from the embedded file.
	desc, err := cmddocs.ParseAndExecute(setupIdentityLongDescription, nil)
	if err == nil {
		cmd.Long = desc.Long
	}

	return cmd
}

func runSetupIdentity(cmd *cobra.Command, _ []string) error {
	presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
	ctx := cmd.Context()

	presenter.Summary("Secure Environment Bootstrap")

	// --- Phase 1: Plumbing (Configuration) ---
	presenter.Header("1. Configuring Tools & Shell")

	if err := configureGPGAgent(ctx, presenter); err != nil {
		return err
	}

	if err := configureGitSecurity(ctx, presenter); err != nil {
		return err
	}

	if err := configureBashrc(presenter); err != nil {
		return err
	}

	presenter.Newline()

	// --- Phase 2: Identity (Interactive) ---
	presenter.Header("2. Identity & Secrets")

	keyID, err := importGPGKey(ctx, presenter)
	if err != nil {
		return err
	}

	if err := trustGPGKey(ctx, presenter, keyID); err != nil {
		return err
	}

	if err := initPass(ctx, presenter, keyID); err != nil {
		return err
	}

	if err := authenticateGitHub(ctx, presenter); err != nil {
		return err
	}

	presenter.Success("Bootstrap Complete! Your environment is secure.")
	presenter.Advice("Run 'source ~/.bashrc' to refresh your shell configuration.")

	return nil
}

func configureGPGAgent(ctx context.Context, presenter *ui.Presenter) (err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	gnupgDir := filepath.Join(home, ".gnupg")
	if err = os.MkdirAll(gnupgDir, dirPermSecure); err != nil {
		return fmt.Errorf("failed to create ~/.gnupg: %w", err)
	}

	// Find pinentry-curses
	out, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "which", "pinentry-curses")
	if err != nil {
		presenter.Warning("pinentry-curses not found. GPG signing might fail in terminal.")
	}

	pinentryPath := strings.TrimSpace(out)
	if pinentryPath == "" {
		return nil
	}

	confPath := filepath.Join(gnupgDir, "gpg-agent.conf")
	newLine := fmt.Sprintf("pinentry-program %s\n", pinentryPath)

	content, err := os.ReadFile(confPath)
	if err == nil {
		if strings.Contains(string(content), "pinentry-program") {
			presenter.Info("GPG Agent already configured.")

			return nil
		}
	}

	configFile, err := os.OpenFile(confPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermRW)
	if err != nil {
		return fmt.Errorf("failed to open gpg-agent.conf: %w", err)
	}

	defer func() {
		err = errors.Join(err, configFile.Close())
	}()

	if _, err := configFile.WriteString(newLine); err != nil {
		return fmt.Errorf("failed to write gpg-agent.conf: %w", err)
	}

	if err := globals.ExecClient.Execute(ctx, ".", "gpg-connect-agent", "reloadagent", "/bye"); err != nil {
		presenter.Warning("failed to reload GPG agent; a manual reload may be needed")
	}

	presenter.Success("✓ GPG Agent configured with %s", pinentryPath)

	return nil
}

func configureGitSecurity(ctx context.Context, presenter *ui.Presenter) error {
	keyID := os.Getenv("GPG_KEY_ID")
	if keyID == "" {
		presenter.Info("GPG_KEY_ID env var not set. Skipping automatic Git signing config.")

		return nil
	}

	cmds := [][]string{
		{"config", "--global", "user.signingkey", keyID},
		{"config", "--global", "commit.gpgsign", "true"},
		{"config", "--global", "gpg.program", "gpg"},
	}

	for _, args := range cmds {
		if err := globals.ExecClient.Execute(ctx, ".", "git", args...); err != nil {
			return fmt.Errorf("failed to configure git: %w", err)
		}
	}

	presenter.Success("✓ Git configured to sign commits with key %s", keyID)

	return nil
}

func configureBashrc(presenter *ui.Presenter) (err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	bashrcPath := filepath.Join(home, ".bashrc")
	marker := "# --- SECURE ENV CONFIG ---"

	content, err := os.ReadFile(bashrcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .bashrc: %w", err)
	}

	if strings.Contains(string(content), marker) {
		presenter.Info("Shell configuration already present.")

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

	file, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermRead)
	if err != nil {
		return fmt.Errorf("failed to open .bashrc: %w", err)
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	if _, err = file.WriteString(block); err != nil {
		return fmt.Errorf("failed to write to .bashrc: %w", err)
	}

	presenter.Success("✓ Shell configuration updated (.bashrc)")

	return nil
}

func importGPGKey(ctx context.Context, presenter *ui.Presenter) (string, error) {
	out, _, _ := globals.ExecClient.CaptureOutput(ctx, ".", "gpg", "--list-secret-keys", "--with-colons")
	if strings.Contains(out, "sec:") {
		presenter.Info("Secret key already exists. Skipping import.")

		return extractKeyID(out), nil
	}

	presenter.Info("👉 Please paste your ASCII-Armored Private GPG Key.")
	presenter.Info("   (Press Enter, paste key, then press Ctrl+D to finish)")

	if err := globals.ExecClient.Execute(ctx, ".", "gpg", "--import"); err != nil {
		return "", fmt.Errorf("gpg import failed: %w", err)
	}

	out, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "gpg", "--list-secret-keys", "--with-colons")
	if err != nil {
		return "", fmt.Errorf("failed to list keys after import: %w", err)
	}

	keyID := extractKeyID(out)
	if keyID == "" {
		return "", ErrNoSecretKey
	}

	return keyID, nil
}

func extractKeyID(gpgOutput string) string {
	for line := range strings.SplitSeq(gpgOutput, "\n") {
		if strings.HasPrefix(line, "sec:") {
			parts := strings.Split(line, ":")
			if len(parts) >= minKeyParts {
				return parts[4]
			}
		}
	}

	return ""
}

func trustGPGKey(ctx context.Context, presenter *ui.Presenter, keyID string) error {
	presenter.Step("Applying 'Ultimate Trust' to key: %s", keyID)

	out, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "gpg", "--list-keys", "--with-colons", keyID)
	if err != nil {
		return fmt.Errorf("failed to get key details: %w", err)
	}

	var fingerprint string

	for line := range strings.SplitSeq(out, "\n") {
		if strings.HasPrefix(line, "fpr:") {
			parts := strings.Split(line, ":")
			if len(parts) >= fingerprintMinParts {
				fingerprint = parts[fingerprintPartIdx]

				break
			}
		}
	}

	if fingerprint == "" {
		return fmt.Errorf("%w: %s", ErrFingerprintNotFound, keyID)
	}

	trustData := fingerprint + ":6:\n"

	if err := runWithStdin(ctx, ".", trustData, "gpg", "--import-ownertrust"); err != nil {
		presenter.Warning("Failed to automate trust setting. You may need to trust the key manually.")
	} else {
		presenter.Success("✓ Key trusted.")
	}

	return nil
}

func initPass(ctx context.Context, presenter *ui.Presenter, keyID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	passDir := filepath.Join(home, ".password-store")

	_, err = os.Stat(passDir)
	if err == nil {
		presenter.Info("Password store already initialized.")

		return nil
	}

	presenter.Step("Initializing 'pass' vault...")

	if err := globals.ExecClient.Execute(ctx, ".", "pass", "init", keyID); err != nil {
		return fmt.Errorf("pass init failed: %w", err)
	}

	presenter.Success("✓ Vault initialized.")

	return nil
}

func authenticateGitHub(ctx context.Context, presenter *ui.Presenter) error {
	var token string

	presenter.Newline()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub Personal Access Token (Fine-grained)").
				Description("Paste your token here. It will be stored securely in 'pass'.").
				EchoMode(huh.EchoModePassword).
				Value(&token),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("input form failed: %w", err)
	}

	if strings.TrimSpace(token) == "" {
		return ErrEmptyToken
	}

	presenter.Step("Storing token in vault...")

	if err := runWithStdin(ctx, ".", token+"\n", "pass", "insert", "-m", "-f", "github/token"); err != nil {
		return fmt.Errorf("failed to store token in pass: %w", err)
	}

	presenter.Success("✓ Token stored in vault (github/token).")

	presenter.Step("Authenticating GitHub CLI...")

	if err := runWithStdin(ctx, ".", token, "gh", "auth", "login", "--with-token"); err != nil {
		return fmt.Errorf("gh auth login failed: %w", err)
	}

	if err := globals.ExecClient.Execute(ctx, ".", "gh", "auth", "setup-git"); err != nil {
		presenter.Warning("failed to set up git authentication with gh; manual setup may be needed")
	}

	presenter.Success("✓ GitHub CLI authenticated.")

	return nil
}

func runWithStdin(ctx context.Context, dir, input, name string, args ...string) error {
	cmd := osexec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s failed: %w", name, err)
	}

	return nil
}
