// Package main implements the ContextVibes bootstrap tool.
// It installs the CLI binary and then hands over control to the CLI
// to finish configuring the environment.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Configuration Defaults.
const (
	defaultRepoURL = "https://github.com/contextvibes/cli.git"
	defaultBranch  = "main"
	targetPath     = "cmd/contextvibes"
	tempDir        = "contextvibes-install-temp"
)

// ANSI Colors.
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
)

func main() {
	forceInstall := flag.Bool("force", false, "Force reinstall of the binary")
	branchName := flag.String("branch", defaultBranch, "Git branch to install from")

	flag.Parse()

	defer cleanup()

	ctx := context.Background()
	goPath := getGoPath(ctx)
	goBin := filepath.Join(goPath, "bin")
	binaryName := "contextvibes"
	binaryPath := filepath.Join(goBin, binaryName)

	// --- Phase 1: The Bootstrap (Get the Binary) ---
	if *forceInstall || !isBinaryInstalled(binaryName) {
		printInfo("Bootstrapping ContextVibes...")
		checkGoInstalled()

		printInfo(fmt.Sprintf("Cloning branch '%s'...", *branchName))
		cloneRepo(ctx, *branchName)

		printInfo("Building and Installing Binary...")
		installBinary(ctx, goBin)
	} else {
		printInfo("ContextVibes binary already installed. Skipping build.")
	}

	// Ensure we can run it
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		logError("Binary not found at expected location: " + binaryPath)
	}

	// --- Phase 2: The Handover (CLI Takes Over) ---

	// 1. Install Tools & Configure Shell
	// We call the CLI to handle the dev tools (govulncheck, etc) and .bashrc
	printInfo("Handing over: Installing Dev Tools & Shell Config...")

	if err := runCommand(ctx, binaryPath, "factory", "tools", "--yes"); err != nil {
		printDetail(fmt.Sprintf("! Warning: 'factory tools' failed: %v", err), colorYellow)
	}

	// 2. Scaffold Environment (VS Code, IDX)
	// We call the CLI to generate settings.json and .idx files
	printInfo("Handing over: Scaffolding Environment...")
	// We assume 'idx' target for now, or we could detect the environment
	if err := runCommand(ctx, binaryPath, "factory", "scaffold", "idx", "--yes"); err != nil {
		printDetail(fmt.Sprintf("! Warning: 'factory scaffold' failed: %v", err), colorYellow)
	}

	// 3. Setup Identity (Git, GPG, Pass)
	// We call the CLI to handle the secure identity setup
	printInfo("Handing over: Identity Setup...")
	// We do NOT use --yes here to allow interactive prompts for GPG keys
	if err := runCommand(ctx, binaryPath, "factory", "setup-identity"); err != nil {
		printDetail("! Identity setup skipped or failed.", colorYellow)
	}

	printSuccess(goBin)
}

// --- Helpers ---

func isBinaryInstalled(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

func checkGoInstalled() {
	if _, err := exec.LookPath("go"); err != nil {
		logError("Go is not installed or not in PATH.")
	}
}

func getGoPath(ctx context.Context) string {
	if gp := os.Getenv("GOPATH"); gp != "" {
		return gp
	}

	cmd := exec.CommandContext(ctx, "go", "env", "GOPATH")

	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}

	home, _ := os.UserHomeDir()

	return filepath.Join(home, "go")
}

func cloneRepo(ctx context.Context, branchName string) {
	_ = os.RemoveAll(tempDir)
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", branchName, defaultRepoURL, tempDir)
	cmd.Stdout = nil

	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logError(fmt.Sprintf("Failed to clone repository (branch: %s).", branchName))
	}
}

func installBinary(ctx context.Context, goBin string) {
	cmd := exec.CommandContext(ctx, "go", "install", ".")
	cmd.Dir = filepath.Join(tempDir, targetPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOBIN="+goBin)
	cmd.Stdout = os.Stdout

	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logError("Build failed.")
	}
}

func runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Stdin = os.Stdin // Connect stdin for interactive commands (setup-identity)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func cleanup() {
	if _, err := os.Stat(tempDir); err == nil {
		_ = os.RemoveAll(tempDir)
	}
}

// --- Output Helpers ---

//nolint:forbidigo
func printInfo(msg string) {
	fmt.Printf("%s==>%s %s%s%s\n", colorBlue, colorReset, colorBold, msg, colorReset)
}

//nolint:forbidigo
func printDetail(msg string, colorCode string) {
	fmt.Printf("  %s%s%s\n", colorCode, msg, colorReset)
}

//nolint:forbidigo
func logError(msg string) {
	fmt.Printf("%sERROR:%s %s\n", colorRed, colorReset, msg)
	cleanup()
	os.Exit(1)
}

//nolint:forbidigo
func printSuccess(goBin string) {
	fmt.Println("")
	fmt.Printf("%s✔ Bootstrap Complete!%s\n", colorGreen, colorReset)
	fmt.Println("------------------------------------------------")
	fmt.Printf("Binary location: %s%s/contextvibes%s\n", colorBold, goBin, colorReset)

	// Check if the binary is in the current PATH
	path, err := exec.LookPath("contextvibes")
	if err == nil && strings.HasPrefix(path, goBin) {
		fmt.Printf("Status:          %sActive and ready.%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("Status:          %sInstalled, but shell restart required.%s\n", colorYellow, colorReset)
		fmt.Println("Please run:      source ~/.bashrc")
	}

	fmt.Println("------------------------------------------------")
}
