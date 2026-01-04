package quality

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/config/assets"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	// aiContextFile is the local constant for the AI context file.
	aiContextFile = "_contextvibes.md"

	// Quality check modes.
	modeEssential  = "essential"
	modeStrict     = "strict"
	modeStyle      = "style"
	modeComplexity = "complexity"
	modeSecurity   = "security"
	modeLocal      = "local"
)

// CheckResult holds the outcome of a single quality check.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
	Err     error  `json:"-"`
}

// NewQualityCmd creates the quality command and its subcommands.
func NewQualityCmd() *cobra.Command {
	var qualityMode string

	cmd := &cobra.Command{
		Use:   "quality [paths...]",
		Args:  cobra.ArbitraryArgs,
		Short: "Runs a series of quality checks against the codebase.",
		Long: `The quality command runs a configurable pipeline of code quality checks.

MODES:
  - essential:  (Default) Basic sanity checks (build, vet, vuln). Fast & recommended for local dev.
  - strict:     Enforces strict linting rules using the embedded 'strict' configuration.
  - style:      Focuses purely on code style, formatting, and naming conventions.
  - complexity: Checks for cyclomatic complexity and function length.
  - security:   Deep security scan (gitleaks, gosec).
  - local:      Uses the project's own .golangci.yml configuration (if present).

EXAMPLES:
  contextvibes product quality                  # Run essential checks on whole project
  contextvibes product quality --mode strict    # Run strict checks
  contextvibes product quality cmd/factory      # Run checks only on specific package`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuality(cmd, args, qualityMode)
		},
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		SilenceErrors:     true,
	}

	usage := fmt.Sprintf("Quality check mode (%s)", strings.Join(supportedModesAsString(), "|"))
	cmd.Flags().StringVarP(&qualityMode, "mode", "m", modeEssential, usage)

	_ = cmd.RegisterFlagCompletionFunc("mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return supportedModesAsString(), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.AddCommand(serveCmd)

	return cmd
}

func runQuality(cmd *cobra.Command, args []string, qualityMode string) error {
	presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
	ctx := cmd.Context()

	// Validate the mode
	isValidMode := false

	for _, validMode := range supportedModesAsString() {
		if qualityMode == validMode {
			isValidMode = true

			break
		}
	}

	if !isValidMode {
		presenter.Error("Invalid quality mode specified: " + qualityMode)

		return fmt.Errorf("invalid quality mode: %s", qualityMode)
	}

	_, err := RunQualityChecks(ctx, presenter, qualityMode, args)

	return err
}

// RunQualityChecks executes the full pipeline of code quality checks.
//
//nolint:cyclop,funlen // Orchestration logic requires complexity.
func RunQualityChecks(
	ctx context.Context,
	presenter *ui.Presenter,
	mode string,
	args []string,
) ([]CheckResult, error) {
	presenter.Header("--- Code Quality Pipeline ---")
	presenter.Info("Mode: %s", mode)

	results := []CheckResult{}

	// 1. Determine which checks to run based on mode
	runBuild := mode == modeEssential || mode == modeStrict || mode == modeLocal
	runVet := mode == modeEssential || mode == modeStrict || mode == modeLocal
	runVuln := mode == modeEssential || mode == modeStrict || mode == modeSecurity || mode == modeLocal
	runGitleaks := mode == modeSecurity

	// Determine Linter Config
	// FIX: Use config.AssetType here to match config.AssetLintStrict etc.
	var linterConfig config.AssetType

	runLint := true

	switch mode {
	case modeEssential:
		runLint = false // Essential is just build/vet/vuln
	case modeStrict:
		linterConfig = config.AssetLintStrict
	case modeStyle:
		linterConfig = config.AssetLintStyle
	case modeComplexity:
		linterConfig = config.AssetLintComplexity
	case modeSecurity:
		linterConfig = config.AssetLintSecurity
	case modeLocal:
		linterConfig = "" // Empty means use local file
	}

	// --- Check: Go Compiler (Build) ---
	if runBuild {
		presenter.Step("Running check: Go Compiler (Build)...")

		_, stderr, err := globals.ExecClient.CaptureOutput(ctx, ".", "go", "build", "./...")
		check := CheckResult{Name: "Go Build", Passed: err == nil, Details: stderr, Err: err}

		results = append(results, check)
		if check.Passed {
			presenter.Success("✓ Code compiles successfully")
		} else {
			presenter.Error("✗ Build failed")
			// If build fails, usually no point continuing
			return finishPipeline(ctx, presenter, results)
		}
	}

	// --- Check: Go Vet ---
	if runVet {
		presenter.Step("Running check: Go Vet...")

		_, stderr, err := globals.ExecClient.CaptureOutput(ctx, ".", "go", "vet", "./...")
		check := CheckResult{Name: "Go Vet", Passed: err == nil, Details: stderr, Err: err}

		results = append(results, check)
		if check.Passed {
			presenter.Success("✓ Code passes go vet")
		} else {
			presenter.Error("✗ Vet found issues")
		}
	}

	// --- Check: GolangCI-Lint ---
	if runLint {
		linterName := "GolangCI-Lint"
		if linterConfig != "" {
			linterName += fmt.Sprintf(" (%s)", linterConfig)
		} else {
			linterName += " (local)"
		}

		presenter.Step("Running check: %s...", linterName)

		linterArgs := []string{"run"}
		if len(args) > 0 {
			linterArgs = append(linterArgs, args...)
		}

		var (
			configFile string
			cleanup    func()
		)

		// If using an embedded config, write it to a temp file

		if linterConfig != "" {
			var err error
			// FIX: Pass config.AssetType to createTempLintConfig
			configFile, cleanup, err = createTempLintConfig(linterConfig)
			if err != nil {
				presenter.Error("Failed to prepare linter config: %v", err)

				return results, err
			}
			defer cleanup()

			linterArgs = append(linterArgs, "-c", configFile)
		}

		stdout, stderr, err := globals.ExecClient.CaptureOutput(ctx, ".", "golangci-lint", linterArgs...)

		// Clean up temp file path from output for readability
		if configFile != "" {
			stderr = strings.ReplaceAll(stderr, configFile, "embedded-config.yml")
			stdout = strings.ReplaceAll(stdout, configFile, "embedded-config.yml")
		}

		check := CheckResult{Name: linterName, Passed: err == nil, Details: stdout + "\n" + stderr, Err: err}
		results = append(results, check)

		if check.Passed {
			presenter.Success("✓ Linter found no issues")
		} else {
			presenter.Error("✗ Linter found issues")
		}
	}

	// --- Check: Go Vulnerability Check ---
	if runVuln {
		presenter.Step("Running check: Go Vulnerability Check...")

		_, stderr, err := globals.ExecClient.CaptureOutput(ctx, ".", "govulncheck", "./...")
		check := CheckResult{Name: "Go Vulnerability Check", Passed: err == nil, Details: stderr, Err: err}

		results = append(results, check)
		if check.Passed {
			presenter.Success("✓ No known vulnerabilities found")
		} else {
			presenter.Error("✗ Vulnerability check failed")
		}
	}

	// --- Check: Gitleaks ---
	if runGitleaks {
		presenter.Step("Running check: Secret Scanning (gitleaks)...")

		if globals.ExecClient.CommandExists("gitleaks") {
			stdout, stderr, err := globals.ExecClient.CaptureOutput(ctx, ".", "gitleaks", "detect", "--no-git", "--verbose")
			check := CheckResult{Name: "Gitleaks", Passed: err == nil, Details: stdout + "\n" + stderr, Err: err}

			results = append(results, check)
			if check.Passed {
				presenter.Success("✓ No secrets detected")
			} else {
				presenter.Error("✗ Secrets detected")
			}
		} else {
			presenter.Warning("gitleaks not found, skipping.")
		}
	}

	return finishPipeline(ctx, presenter, results)
}

func finishPipeline(ctx context.Context, presenter *ui.Presenter, results []CheckResult) ([]CheckResult, error) {
	presenter.Header("--- Pipeline Summary ---")

	failedChecks := 0

	for _, r := range results {
		if !r.Passed {
			failedChecks++

			presenter.Error("  - %s: %s", r.Name, color.RedString("FAILED"))
		} else {
			presenter.Success("  - %s: PASSED", r.Name)
		}
	}

	presenter.Newline()

	if failedChecks > 0 {
		if err := generateAIContextFile(ctx, results); err != nil {
			presenter.Warning("Failed to generate AI context file: %v", err)
		} else {
			presenter.Info("Generated AI Context: %s", aiContextFile)
			presenter.Advice("Pass this file to your AI to fix the issues.")
		}

		presenter.Error("  %d check(s) failed.", failedChecks)

		return results, errors.New("one or more quality checks failed")
	}

	if err := removeAIContextFile(); err == nil {
		presenter.Info("Removed stale AI Context file: %s (all checks passed)", aiContextFile)
	}

	presenter.Success("  All checks passed!")

	return results, nil
}

// FIX: Use config.AssetType as parameter type.
func createTempLintConfig(assetType config.AssetType) (string, func(), error) {
	configBytes, err := config.GetLanguageAsset("go", assetType)
	if err != nil {
		return "", nil, fmt.Errorf("failed to load config asset: %w", err)
	}

	tmpFile, err := os.CreateTemp(".", ".golangci-"+string(assetType)+"-*.yml")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp config: %w", err)
	}

	cleanup := func() {
		_ = os.Remove(tmpFile.Name())
	}

	if _, err := tmpFile.Write(configBytes); err != nil {
		_ = tmpFile.Close()

		cleanup()

		return "", nil, fmt.Errorf("failed to write config: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		cleanup()

		return "", nil, fmt.Errorf("failed to close config file: %w", err)
	}

	return tmpFile.Name(), cleanup, nil
}

func generateAIContextFile(ctx context.Context, results []CheckResult) error {
	// Use assets package for the AI prompt template
	headerTmpl, err := assets.GetLanguageAsset("go", assets.AIContextPrompt)
	if err != nil {
		return fmt.Errorf("failed to get AI context header template: %w", err)
	}

	var builder strings.Builder
	builder.Write(headerTmpl)

	for _, r := range results {
		if !r.Passed {
			builder.WriteString(fmt.Sprintf("\n### ❌ %s\n", r.Name))

			if r.Details != "" {
				builder.WriteString(fmt.Sprintf("\n```\n%s\n```\n", strings.TrimSpace(r.Details)))
			}
		}
	}

	//nolint:mnd,gosec // 0644 is a standard file permission.
	if err := os.WriteFile(aiContextFile, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write AI context file: %w", err)
	}

	return nil
}

func removeAIContextFile() error {
	if _, err := os.Stat(aiContextFile); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(aiContextFile)
}

func supportedModesAsString() []string {
	return []string{modeEssential, modeStrict, modeStyle, modeComplexity, modeSecurity, modeLocal}
}
