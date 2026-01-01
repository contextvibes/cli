// Package scaffold provides the command to generate infrastructure configuration.
package scaffold

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed scaffold.md.tpl
var scaffoldLongDescription string

// ScaffoldCmd represents the scaffold command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ScaffoldCmd = &cobra.Command{
	Use:   "scaffold [target]",
	Short: "Scaffolds infrastructure (e.g., idx, firebase).",
	Example: `  contextvibes factory scaffold idx
  contextvibes factory scaffold firebase`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()
		target := args[0]

		switch target {
		case "idx":
			return scaffoldIDX(presenter)
		case "firebase":
			return scaffoldFirebase(ctx, presenter)
		default:
			//nolint:err113 // Dynamic error is appropriate for CLI output.
			return fmt.Errorf("unsupported scaffold target: %s (supported: idx, firebase)", target)
		}
	},
}

func scaffoldIDX(presenter *ui.Presenter) error {
	presenter.Summary("Scaffolding Project IDX Environment (.idx/)")

	idxDir := ".idx"
	//nolint:mnd // 0750 is standard dir permission.
	if err := os.MkdirAll(idxDir, 0o750); err != nil {
		return fmt.Errorf("failed to create .idx directory: %w", err)
	}

	files := map[string]string{
		"dev.nix":           devNixTemplate,
		"contextvibes.nix":  contextvibesNixTemplate,
		"golangci-lint.nix": golangciLintNixTemplate,
		"local.nix":         localNixTemplate,
	}

	for filename, content := range files {
		path := filepath.Join(idxDir, filename)

		// Check existence to avoid accidental overwrite
		if _, err := os.Stat(path); err == nil {
			presenter.Info("  ~ %s already exists. Skipping.", filename)

			continue
		}

		//nolint:mnd // 0600 is standard secure file permission.
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		presenter.Success("  + Created %s", filename)
	}

	presenter.Newline()
	presenter.Advice("Environment scaffolded. You may need to rebuild your environment for changes to take effect.")
	presenter.Advice("Edit .idx/local.nix to set your GPG_KEY_ID.")

	return nil
}

func scaffoldFirebase(ctx context.Context, presenter *ui.Presenter) error {
	presenter.Summary("Scaffolding Firebase Environment")

	// 1. Check for Tooling
	if !globals.ExecClient.CommandExists("firebase") {
		presenter.Error("Firebase CLI not found.")
		presenter.Advice("Please rebuild your environment (dev.nix) to include 'firebase-tools'.")
		//nolint:err113 // Dynamic error is appropriate for CLI output.
		return errors.New("firebase-tools missing")
	}

	// 2. Login Check (Simple heuristic)
	// We run a harmless command to check auth state
	_, _, err := globals.ExecClient.CaptureOutput(ctx, ".", "firebase", "projects:list", "--json")
	if err != nil {
		presenter.Warning("You do not appear to be logged in to Firebase.")
		presenter.Step("Running 'firebase login'...")
		// Interactive login
		err = globals.ExecClient.Execute(ctx, ".", "firebase", "login")
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	// 3. Init
	presenter.Step("Initializing Firebase Project Structure...")
	presenter.Info("This will guide you through creating firebase.json and .firebaserc")

	// We use interactive execution here because firebase init is highly interactive
	err = globals.ExecClient.Execute(ctx, ".", "firebase", "init")
	if err != nil {
		return fmt.Errorf("firebase init failed: %w", err)
	}

	presenter.Success("Firebase environment scaffolded successfully.")

	return nil
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(scaffoldLongDescription, nil)
	if err != nil {
		panic(err)
	}

	ScaffoldCmd.Short = desc.Short
	ScaffoldCmd.Long = desc.Long
}
