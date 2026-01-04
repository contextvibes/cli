package format

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

const filePermUserRW = 0o600

//go:embed format.md.tpl
var formatLongDescription string

var (
	errFormattingFailed = errors.New("one or more formatting tools failed")
	nolintRegex         = regexp.MustCompile(`//\s*nolint.*`)
)

var (
	removeDirectives bool
	formatMode       string
)

// FormatCmd represents the format command.
var FormatCmd = &cobra.Command{
	Use:   "format [paths...]",
	Short: "Applies code formatting and auto-fixes linter issues.",
	Long: `Applies standard formatting and auto-fixable linter suggestions.

MODES:
  - essential: (Default) Runs standard tools (gofmt, goimports, go mod tidy). Safe and fast.
  - style:     Runs essential + fixes style issues (revive, whitespace, etc).
  - strict:    Runs essential + fixes strict linting issues.
  - local:     Runs essential + fixes using your local .golangci.yml.
  - security:  (Same as essential - security tools rarely auto-fix).
  - complexity:(Same as essential - complexity tools rarely auto-fix).`,
	Example: `  contextvibes product format                  # Run essential formatting
  contextvibes product format --mode style     # Fix style issues
  contextvibes product format --mode local     # Use local config
  contextvibes product format cmd/factory      # Format specific package`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		// Validate mode
		isValid := false
		for _, m := range supportedModesAsString() {
			if formatMode == m {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid mode: %s", formatMode)
		}

		presenter.Summary("Applying code formatting and auto-fixes")
		presenter.Info("Mode: %s", formatMode)

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		gitClient, err := initGitClient(ctx, cwd)
		if err != nil {
			presenter.Warning("Could not initialize Git client. Safety checks disabled.")
		}

		projType, err := project.Detect(cwd)
		if err != nil {
			return fmt.Errorf("failed to detect project type: %w", err)
		}

		var reportBuffer bytes.Buffer
		initReportBuffer(&reportBuffer)

		formatErrors := runFormatters(ctx, presenter, &reportBuffer, projType, gitClient, args)

		presenter.Newline()
		if len(formatErrors) > 0 {
			return errFormattingFailed
		}

		presenter.Success("All formatting tools completed.")

		if err := dumpUpdatedFiles(ctx, gitClient, &reportBuffer, args); err != nil {
			presenter.Warning("Failed to dump updated file content: %v", err)
		}

		if err := tools.WriteBufferToFile("_contextvibes.md", &reportBuffer); err != nil {
			presenter.Error("Failed to write report: %v", err)
		} else {
			presenter.Success("Full context written to %s", presenter.Highlight("_contextvibes.md"))
		}

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func initGitClient(ctx context.Context, cwd string) (*git.GitClient, error) {
	gitCfg := git.GitClientConfig{
		Logger:   globals.AppLogger,
		Executor: globals.ExecClient.UnderlyingExecutor(),
	}
	return git.NewClient(ctx, cwd, gitCfg)
}

func initReportBuffer(buf *bytes.Buffer) {
	_, _ = fmt.Fprintf(buf, "# Format Report (%s)\n\n", time.Now().Format(time.RFC3339))
	_, _ = fmt.Fprintf(buf, "## Execution Log\n\n")
}

func runFormatters(ctx context.Context, presenter *ui.Presenter, buf *bytes.Buffer, projType project.Type, gitClient *git.GitClient, args []string) []error {
	var formatErrors []error
	cwd, _ := os.Getwd()

	switch projType {
	case project.Go:
		if removeDirectives {
			modified, err := runStripDirectives(ctx, presenter, gitClient, args)
			if err != nil {
				formatErrors = append(formatErrors, err)
				return formatErrors
			}
			logStrippedFiles(buf, modified)
		}

		runGoFormatters(ctx, presenter, buf, cwd, &formatErrors, args)
	default:
		presenter.Info("No formatters configured for %s", projType)
	}

	return formatErrors
}

func runGoFormatters(ctx context.Context, presenter *ui.Presenter, buf *bytes.Buffer, cwd string, formatErrors *[]error, args []string) {
	// 1. Standard Tools (Always Run)
	if globals.ExecClient.CommandExists("go") {
		out, err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "go", []string{"mod", "tidy"})
		logOutput(buf, "go mod tidy", out, err)
		if err != nil {
			*formatErrors = append(*formatErrors, err)
		} else {
			presenter.Success("✓ go mod tidy applied.")
		}
	}

	if globals.ExecClient.CommandExists("goimports") {
		goimportsArgs := append([]string{"-l", "-w"}, args...)
		if len(args) == 0 {
			goimportsArgs = append(goimportsArgs, ".")
		}
		out, err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "goimports", goimportsArgs)
		logOutput(buf, "goimports", out, err)
		if err != nil {
			*formatErrors = append(*formatErrors, err)
		} else {
			logUpdatedFiles(presenter, out)
			presenter.Success("✓ goimports applied.")
		}
	}

	gofmtArgs := append([]string{"-l", "-s", "-w"}, args...)
	if len(args) == 0 {
		gofmtArgs = append(gofmtArgs, ".")
	}
	out, err := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "gofmt", gofmtArgs)
	logOutput(buf, "gofmt", out, err)
	if err != nil {
		*formatErrors = append(*formatErrors, err)
	} else {
		logUpdatedFiles(presenter, out)
		presenter.Success("✓ gofmt -s applied.")
	}

	// 2. Linter Auto-Fixes (Mode Dependent)
	var linterConfig config.AssetType
	runLint := false

	switch formatMode {
	case "style":
		runLint = true
		linterConfig = config.AssetLintStyle
	case "strict":
		runLint = true
		linterConfig = config.AssetLintStrict
	case "local":
		runLint = true
		linterConfig = "" // Use local file
	case "essential", "security", "complexity":
		runLint = false // These modes don't typically have safe auto-fixes
	}

	if runLint {
		lintOut, lintErr := runGolangCiLint(ctx, presenter, cwd, args, linterConfig)
		logOutput(buf, "golangci-lint --fix", lintOut, lintErr)
		if lintErr != nil {
			presenter.Warning("Linter fixes applied, but issues may remain.")
		} else {
			presenter.Success("✓ golangci-lint --fix applied.")
		}
	}
}

func runGolangCiLint(ctx context.Context, presenter *ui.Presenter, cwd string, args []string, configType config.AssetType) (string, error) {
	lintArgs := []string{"run", "--fix"}

	if configType != "" {
		presenter.Info("Using %s configuration for fixes.", configType)
		configPath, cleanup, err := createTempLintConfig(configType)
		if err != nil {
			return "", err
		}
		defer cleanup()
		lintArgs = append(lintArgs, "-c", configPath)
	} else {
		presenter.Info("Using local configuration for fixes.")
	}

	if len(args) > 0 {
		lintArgs = append(lintArgs, args...)
	}

	lintOut, lintErr := runFormatCommand(ctx, presenter, globals.ExecClient, cwd, "golangci-lint", lintArgs)

	// Sanitize output
	if configType != "" {
		lintOut = strings.ReplaceAll(lintOut, ".golangci-"+string(configType), "embedded-config")
	}

	return lintOut, lintErr
}

func createTempLintConfig(assetType config.AssetType) (string, func(), error) {
	configBytes, err := config.GetLanguageAsset("go", assetType)
	if err != nil {
		return "", nil, fmt.Errorf("failed to load config asset: %w", err)
	}

	tmpFile, err := os.CreateTemp(".", ".golangci-"+string(assetType)+"-*.yml")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp config: %w", err)
	}

	cleanup := func() { _ = os.Remove(tmpFile.Name()) }

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

func logStrippedFiles(buf *bytes.Buffer, modified []string) {
	if len(modified) > 0 {
		_, _ = fmt.Fprintf(buf, "### Strip Directives\n")
		for _, f := range modified {
			_, _ = fmt.Fprintf(buf, "- Cleaned: %s\n", f)
		}
		_, _ = fmt.Fprintf(buf, "\n")
	}
}

func logUpdatedFiles(presenter *ui.Presenter, output string) {
	if len(output) > 0 {
		//nolint:modernize // SplitSeq not available in all envs
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			presenter.Detail("Updated: %s", line)
		}
	}
}

func dumpUpdatedFiles(ctx context.Context, gitClient *git.GitClient, buf *bytes.Buffer, args []string) error {
	if gitClient == nil {
		return nil
	}
	statusOut, _, err := gitClient.GetStatusShort(ctx)
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}
	if len(statusOut) == 0 {
		return nil
	}
	dirtyFiles := parseGitStatus(statusOut)
	filesToDump := filterFiles(dirtyFiles, args)

	if len(filesToDump) > 0 {
		_, _ = fmt.Fprintf(buf, "\n## Updated File Content (Ground Truth)\n")
		for _, path := range filesToDump {
			info, err := os.Stat(path)
			if err == nil && !info.IsDir() {
				content, err := os.ReadFile(path)
				if err == nil {
					tools.AppendFileMarkerHeader(buf, path)
					_, _ = buf.Write(content)
					tools.AppendFileMarkerFooter(buf, path)
				}
			}
		}
	}
	return nil
}

func logOutput(buf *bytes.Buffer, tool string, output string, err error) {
	_, _ = fmt.Fprintf(buf, "### %s\n", tool)
	if err != nil {
		_, _ = fmt.Fprintf(buf, "**Status:** Failed\n")
		_, _ = fmt.Fprintf(buf, "**Error:** %v\n", err)
	} else {
		_, _ = fmt.Fprintf(buf, "**Status:** Success\n")
	}
	if len(output) > 0 {
		_, _ = fmt.Fprintf(buf, "\n```text\n%s\n```\n", strings.TrimSpace(output))
	}
	_, _ = fmt.Fprintf(buf, "\n")
}

func parseGitStatus(output string) []string {
	var files []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) >= 2 {
			files = append(files, parts[len(parts)-1])
		}
	}
	return files
}

func filterFiles(candidates, targets []string) []string {
	if len(targets) == 0 {
		return candidates
	}
	var result []string
	for _, c := range candidates {
		for _, t := range targets {
			cleanC := filepath.Clean(c)
			cleanT := filepath.Clean(t)
			if cleanC == cleanT || strings.HasPrefix(cleanC, cleanT+string(os.PathSeparator)) {
				result = append(result, c)
				break
			}
		}
	}
	return result
}

func runStripDirectives(ctx context.Context, presenter *ui.Presenter, client *git.GitClient, args []string) ([]string, error) {
	presenter.Header("Removing Linter Directives")
	if err := ensureGitIsClean(ctx, presenter, client); err != nil {
		return nil, err
	}
	presenter.Warning("This will remove all '//nolint' directives from Go files.")
	if !globals.AssumeYes {
		confirmed, err := presenter.PromptForConfirmation("Are you sure you want to proceed?")
		if err != nil {
			return nil, fmt.Errorf("confirmation failed: %w", err)
		}
		if !confirmed {
			presenter.Info("Skipping directive removal.")
			return nil, nil
		}
	}
	return stripDirectivesFromPaths(presenter, args)
}

func ensureGitIsClean(ctx context.Context, presenter *ui.Presenter, client *git.GitClient) error {
	if client == nil {
		return nil
	}
	isClean, err := client.IsWorkingDirClean(ctx)
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if isClean {
		return nil
	}
	presenter.Warning("Your working directory is dirty (has uncommitted changes).")
	if globals.AssumeYes {
		presenter.Error("Cannot proceed in non-interactive mode with dirty directory.")
		return errors.New("working directory not clean")
	}
	confirmCommit, err := presenter.PromptForConfirmation("Stage and commit all changes now to proceed?")
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}
	if !confirmCommit {
		presenter.Info("Aborted by user.")
		return errors.New("user aborted on dirty directory")
	}
	presenter.Step("Staging and committing changes...")
	if err := client.AddAll(ctx); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	if err := client.Commit(ctx, "chore: Save state before removing directives"); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	presenter.Success("✓ State saved.")
	return nil
}

func stripDirectivesFromPaths(presenter *ui.Presenter, paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	var modifiedFiles []string
	for _, path := range paths {
		err := filepath.WalkDir(path, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if shouldSkipDir(d) {
				return fs.SkipDir
			}
			if !d.IsDir() && strings.HasSuffix(currentPath, ".go") {
				modified, stripErr := stripFile(currentPath)
				if stripErr != nil {
					presenter.Error("Failed to process %s: %v", currentPath, stripErr)
					return nil
				}
				if modified {
					modifiedFiles = append(modifiedFiles, currentPath)
					presenter.Detail("Cleaned %s", currentPath)
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking path %s: %w", path, err)
		}
	}
	presenter.Success("Removed directives from %d files.", len(modifiedFiles))
	return modifiedFiles, nil
}

func shouldSkipDir(d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	name := d.Name()
	return name == "vendor" || name == ".git"
}

func stripFile(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read failed: %w", err)
	}
	original := string(content)
	modified := nolintRegex.ReplaceAllString(original, "")
	if original == modified {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(modified), filePermUserRW); err != nil {
		return false, fmt.Errorf("write failed: %w", err)
	}
	return true, nil
}

func runFormatCommand(ctx context.Context, p *ui.Presenter, execClient *exec.ExecutorClient, cwd, cmd string, args []string) (string, error) {
	p.Step("Running %s %s...", cmd, args[0])
	if !execClient.CommandExists(cmd) {
		p.Warning("'%s' command not found, skipping.", cmd)
		return "", nil
	}
	stdout, stderr, err := execClient.CaptureOutput(ctx, cwd, cmd, args...)
	if err != nil {
		return stdout + stderr, fmt.Errorf("failed to execute %s: %w", cmd, err)
	}
	return stdout + stderr, nil
}

func supportedModesAsString() []string {
	return []string{"essential", "strict", "style", "complexity", "security", "local"}
}

func init() {
	desc, err := cmddocs.ParseAndExecute(formatLongDescription, nil)
	if err != nil {
		panic(err)
	}

	FormatCmd.Short = desc.Short
	FormatCmd.Long = desc.Long

	usage := fmt.Sprintf("Format mode (%s)", strings.Join(supportedModesAsString(), "|"))
	FormatCmd.Flags().StringVarP(&formatMode, "mode", "m", "essential", usage)
	FormatCmd.Flags().BoolVar(&removeDirectives, "remove-directives", false, "Remove all //nolint directives from Go files")

	// Register completion
	_ = FormatCmd.RegisterFlagCompletionFunc("mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return supportedModesAsString(), cobra.ShellCompDirectiveNoFileComp
	})
}
