// Package upgradecli provides the command to update the CLI version in Nix.
package upgradecli

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/github"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed upgradecli.md.tpl
var upgradeCliLongDescription string

// UpgradeCliCmd represents the upgrade-cli command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var UpgradeCliCmd = &cobra.Command{
	Use:   "upgrade-cli",
	Short: "Updates the CLI version definition in .idx/contextvibes.nix",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		nixFilePath := filepath.Join(".idx", "contextvibes.nix")
		if _, err := os.Stat(nixFilePath); os.IsNotExist(err) {
			presenter.Error("File not found: %s", nixFilePath)
			presenter.Advice("Run 'contextvibes factory scaffold idx' first.")
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("nix file missing")
		}

		// 1. Get Latest Version from GitHub
		presenter.Step("Checking for updates...")
		ghClient, err := github.NewClient(ctx, globals.AppLogger, "contextvibes", "cli")
		if err != nil {
			return fmt.Errorf("failed to init github client: %w", err)
		}

		release, _, err := ghClient.Repositories.GetLatestRelease(ctx, "contextvibes", "cli")
		if err != nil {
			return fmt.Errorf("failed to get latest release: %w", err)
		}

		latestVersion := strings.TrimPrefix(release.GetTagName(), "v")
		presenter.Info("Latest version: %s", latestVersion)

		// 2. Read current file

		contentBytes, err := os.ReadFile(nixFilePath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", nixFilePath, err)
		}
		content := string(contentBytes)

		// 3. Check if update needed
		// Supports both old format (version = "...") and new hybrid format (version ? "...")
		versionRegex := regexp.MustCompile(`version [?=]= "([0-9.]+)";`)
		matches := versionRegex.FindStringSubmatch(content)
		//nolint:mnd // Expecting full match + 1 capture group.
		if len(matches) < 2 {
			//nolint:err113 // Dynamic error is appropriate here.
			return fmt.Errorf("could not parse current version in %s", nixFilePath)
		}
		currentVersion := matches[1]

		if currentVersion == latestVersion {
			presenter.Success("Already up to date (%s).", currentVersion)

			return nil
		}

		presenter.Info("Current version: %s. Upgrading to %s...", currentVersion, latestVersion)

		// 4. Calculate Hash (Prefetch)
		downloadURL := fmt.Sprintf("https://github.com/contextvibes/cli/releases/download/v%s/contextvibes", latestVersion)

		presenter.Step("Prefetching hash for %s...", downloadURL)
		if !globals.ExecClient.CommandExists("nix-prefetch-url") {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("nix-prefetch-url not found; are you in the Nix environment?")
		}

		hashOutput, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "nix-prefetch-url", downloadURL)
		if err != nil {
			return fmt.Errorf("failed to prefetch url: %w", err)
		}
		newHash := strings.TrimSpace(hashOutput)

		// 5. Patch the file
		// Replace Version (Handle both ? and =)
		newContent := versionRegex.ReplaceAllString(content, fmt.Sprintf(`version ? "%s";`, latestVersion))

		// Replace Hash
		// Handle old format (sha256 =) and new format (binHash ?)
		if strings.Contains(newContent, "binHash ?") {
			hashRegex := regexp.MustCompile(`binHash \? "[^"]+";`)
			newContent = hashRegex.ReplaceAllString(newContent, fmt.Sprintf(`binHash ? "%s";`, newHash))
		} else {
			hashRegex := regexp.MustCompile(`sha256 = "[^"]+";`)
			newContent = hashRegex.ReplaceAllString(newContent, fmt.Sprintf(`sha256 = "%s";`, newHash))
		}

		// 6. Write back
		//nolint:mnd // 0600 is standard secure file permission.
		if err := os.WriteFile(nixFilePath, []byte(newContent), 0o600); err != nil {
			return fmt.Errorf("failed to write %s: %w", nixFilePath, err)
		}

		presenter.Success("Updated %s to version %s.", nixFilePath, latestVersion)
		presenter.Advice("Please rebuild your environment (Command Palette > Rebuild Environment) to apply changes.")

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(upgradeCliLongDescription, nil)
	if err != nil {
		panic(err)
	}

	UpgradeCliCmd.Short = desc.Short
	UpgradeCliCmd.Long = desc.Long
}
