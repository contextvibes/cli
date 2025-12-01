// Package scrub provides the cleanup functionality for the development environment.
package scrub

import (
	"context"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

const (
	// dirPermUserRWX represents read/write/execute permissions for the user (0750).
	dirPermUserRWX = 0o750
)

// ScrubCmd represents the scrub command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var ScrubCmd = &cobra.Command{
	Use:   "scrub",
	Short: "Deep cleans the development environment (Docker, Nix, Caches).",
	Long: `Performs a "Scorched Earth" cleanup of the development environment.
Targeting:
- Android Emulator & SDK artifacts (.emu, .androidsdkroot)
- Go build and module caches
- Docker system (prune all)
- Nix garbage (old generations)
- General user cache (~/.cache)

WARNING: This is destructive and will remove cached data to free up space.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Warning("INITIATING DEEP CLEANUP PROTOCOL")
		presenter.Warning("This will wipe Docker images, Go caches, and Nix garbage.")

		if !globals.AssumeYes {
			confirmed, err := presenter.PromptForConfirmation("Are you sure you want to proceed?")
			if err != nil || !confirmed {
				presenter.Info("Scrub aborted.")

				return nil
			}
		}

		// 1. Android Artifacts
		cleanAndroid(ctx, presenter)

		// 2. Go Artifacts
		cleanGo(ctx, presenter)

		// 3. Docker Artifacts
		cleanDocker(ctx, presenter)

		// 4. Nix Garbage
		cleanNix(ctx, presenter)

		// 5. General Cache
		cleanUserCache(presenter)

		// 6. Report
		presenter.Header("--- Final Disk Usage ---")
		// We use 'du' here as it's the standard way to check usage in this env
		_ = globals.ExecClient.Execute(ctx, ".", "du", "-h", "--max-depth=1", ".")

		presenter.Success("Cleanup Complete. You are ready to build.")

		return nil
	},
}

func cleanAndroid(ctx context.Context, presenter *ui.Presenter) {
	presenter.Header("1. Cleaning Android Artifacts")

	// The bash script uses 'find -delete' to avoid "Device busy" errors on the root folder.
	// We replicate that behavior by executing find directly.
	dirs := []string{".emu", ".androidsdkroot"}

	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			presenter.Info("   - %s not found.", dir)

			continue
		}

		presenter.Step("Emptying %s folder...", dir)
		// find .emu -mindepth 1 -delete
		err = globals.ExecClient.Execute(ctx, ".", "find", dir, "-mindepth", "1", "-delete")
		if err != nil {
			presenter.Warning("     (Some files in %s might be locked, skipping)", dir)
		} else {
			presenter.Success("     ✓ %s cleared.", dir)
		}
	}
}

func cleanGo(ctx context.Context, presenter *ui.Presenter) {
	presenter.Header("2. Cleaning Go Ecosystem")

	if !globals.ExecClient.CommandExists("go") {
		presenter.Warning("   Go not found, skipping.")

		return
	}

	cmds := [][]string{
		{"clean", "-cache"},
		{"clean", "-testcache"},
		{"clean", "-modcache"},
		{"clean", "-fuzzcache"},
	}

	for _, args := range cmds {
		_ = globals.ExecClient.Execute(ctx, ".", "go", args...)
	}

	// Aggressive manual delete
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, "go", "pkg", "mod"),
		filepath.Join(home, ".cache", "go-build"),
	}

	for _, path := range paths {
		err := os.RemoveAll(path)
		if err == nil {
			presenter.Detail("Removed %s", path)
		}
	}

	presenter.Success("   ✓ Go caches cleared.")
}

func cleanDocker(ctx context.Context, presenter *ui.Presenter) {
	presenter.Header("3. Cleaning Docker System")

	// Check if docker is running
	err := globals.ExecClient.Execute(ctx, ".", "docker", "info")
	if err != nil {
		presenter.Warning("   Docker not running or not found, skipping.")

		return
	}

	// Prune everything
	err = globals.ExecClient.Execute(ctx, ".", "docker", "system", "prune", "-a", "--volumes", "-f")
	if err != nil {
		presenter.Error("Failed to prune docker system: %v", err)
	}

	_ = globals.ExecClient.Execute(ctx, ".", "docker", "builder", "prune", "-a", "-f")

	presenter.Success("   ✓ Docker system pruned.")
}

func cleanNix(ctx context.Context, presenter *ui.Presenter) {
	presenter.Header("4. Cleaning Nix Garbage")

	if !globals.ExecClient.CommandExists("nix-collect-garbage") {
		presenter.Warning("   Nix not found, skipping.")

		return
	}

	err := globals.ExecClient.Execute(ctx, ".", "nix-collect-garbage", "-d")
	if err != nil {
		presenter.Error("Failed to collect nix garbage: %v", err)
	} else {
		presenter.Success("   ✓ Nix generations collected.")
	}
}

func cleanUserCache(presenter *ui.Presenter) {
	presenter.Header("5. Cleaning User Cache")

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache")

	// We don't want to delete the directory itself, just contents,
	// but os.RemoveAll on the dir is usually safe enough for ~/.cache
	// Re-creating it ensures it exists.
	err := os.RemoveAll(cacheDir)
	if err != nil {
		presenter.Warning("Could not fully remove .cache: %v", err)
	}

	_ = os.MkdirAll(cacheDir, dirPermUserRWX)

	presenter.Success("   ✓ ~/.cache emptied.")
}
