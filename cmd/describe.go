// cmd/describe.go
package cmd

import (
	"bytes"
	"context" // Ensure context is imported
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"

	// "github.com/contextvibes/cli/internal/tools" // Should no longer be needed for exec functions
	"github.com/contextvibes/cli/internal/tools" // Keep for non-exec tools like ReadFileContent, markdown helpers for now

	gitignore "github.com/denormal/go-gitignore"
	"github.com/spf13/cobra"
)

const (
	defaultDescribeOutputFile = "contextvibes.md"
	includeExtensionsRegex    = `\.(go|mod|sum|tf|py|yaml|json|md|gitignore|txt|hcl|nix)$|^(Taskfile\.yaml|requirements\.txt|README\.md|\.idx/dev\.nix|\.idx/airules\.md)$`
	maxFileSizeKB             = 500
	excludePathsRegex         = `(^\.git/|^\.terraform/|^\.venv/|^__pycache__/|^\.DS_Store|^\.pytest_cache/|^\.vscode/|\.tfstate|\.tfplan|^secrets?/|\.auto\.tfvars|ai_context\.txt|crash.*\.log|contextvibes\.md)`
	treeIgnorePattern         = ".git|.terraform|.venv|venv|env|__pycache__|.pytest_cache|.DS_Store|.idx|.vscode|*.tfstate*|*.log|ai_context.txt|contextvibes.md|node_modules|build|dist"
)

var criticalFiles = []string{
	"./README.md",
	"./.idx/dev.nix",
	"./.gitignore",
}

var describeOutputFile string

var describeCmd = &cobra.Command{
	Use:   "describe [-o <output_file>]",
	Short: "Generates project context file (default: contextvibes.md).",
	Long: `Gathers project context (user prompt, environment, git status, structure, relevant files)
and writes it to a Markdown file (default: ` + defaultDescribeOutputFile + `), suitable for AI interaction.

Respects .gitignore, .aiexclude rules, and file size limits when including file content.`,
	Example: `  contextvibes describe                 # Prompts for input, saves context to contextvibes.md
  contextvibes describe -o project_snapshot.md # Saves context to custom file`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if ExecClient == nil {
			return fmt.Errorf("internal error: executor client not initialized")
		}
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background() // Define context for helper functions

		presenter.Summary("Generating project context description.")

		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Describe: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}

		// Initialize GitClientConfig correctly
		gitCfg := git.GitClientConfig{
			Logger:                logger, // This is AppLogger
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch,
			// DO NOT set Executor here like: Executor: ExecClient.executor
			// Let GitClient's NewClient -> validateAndSetDefaults handle it.
			// It will create an OSCommandExecutor using the provided logger.
		}
		if LoadedAppConfig != nil && LoadedAppConfig.Git.DefaultRemote != "" {
			gitCfg.DefaultRemoteName = LoadedAppConfig.Git.DefaultRemote
		}
		if LoadedAppConfig != nil && LoadedAppConfig.Git.DefaultMainBranch != "" {
			gitCfg.DefaultMainBranchName = LoadedAppConfig.Git.DefaultMainBranch
		}

		logger.DebugContext(ctx, "Initializing GitClient for describe", slog.String("source_command", "describe"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init (is this a Git repository?): %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized for describe", slog.String("source_command", "describe"))
		cwd := client.Path()

		includeRe, err := regexp.Compile(includeExtensionsRegex)
		if err != nil { /* ... */
			return err
		}
		excludeRe, err := regexp.Compile(excludePathsRegex)
		if err != nil { /* ... */
			return err
		}
		maxSizeBytes := int64(maxFileSizeKB * 1024)

		var aiExcluder gitignore.GitIgnore
		aiExcludeFilePath := filepath.Join(cwd, ".aiexclude")
		aiExcludeContent, readErr := os.ReadFile(aiExcludeFilePath)
		gitignoreErrorHandler := func(ignoreErr gitignore.Error) bool {
			presenter.Warning("Parsing .aiexclude file: %v", ignoreErr)
			logger.WarnContext(ctx, "Parsing .aiexclude file", slog.Any("error", ignoreErr))
			return true
		}
		if readErr == nil {
			aiExcluder = gitignore.New(bytes.NewReader(aiExcludeContent), cwd, gitignoreErrorHandler)
			if aiExcluder != nil {
				presenter.Info("Loaded exclusion rules from %s", presenter.Highlight(".aiexclude"))
			} else if len(bytes.TrimSpace(aiExcludeContent)) > 0 {
				presenter.Info(".aiexclude file found but contains no active rules.")
			}
		} else if !os.IsNotExist(readErr) {
			presenter.Warning("Could not read .aiexclude file at %s: %v", aiExcludeFilePath, readErr)
		}

		var outputBuffer bytes.Buffer
		outputFilePath := describeOutputFile
		if outputFilePath == "" {
			outputFilePath = defaultDescribeOutputFile
		}
		presenter.Info("Generating context file: %s", presenter.Highlight(outputFilePath))

		tools.AppendSectionHeader(&outputBuffer, "Prompt")
		presenter.Separator()
		presenter.Step("Enter a prompt for the AI (e.g., 'Refactor X module', 'Add Y feature to script').")
		presenter.Step("Be specific about goals, files, resources, or errors.")
		presenter.Separator()
		userPrompt, err := presenter.PromptForInput("> Prompt: ")
		if err != nil {
			return err
		}
		if userPrompt == "" {
			presenter.Error("prompt cannot be empty")
			return errors.New("prompt cannot be empty")
		}
		fmt.Fprintf(&outputBuffer, "%s\n\n", userPrompt)

		tools.AppendSectionHeader(&outputBuffer, "Collaboration Notes")
		outputBuffer.WriteString("For future reviews:\n")
		outputBuffer.WriteString("- If code changes are significant or span multiple areas, please provide the full updated file(s) using this task.\n")
		outputBuffer.WriteString("- If changes are small and localized (e.g., fixing a typo, a few lines in one function), you can provide just the relevant snippet, but clearly state the filename and function/context.\n")
		outputBuffer.WriteString("- Always describe the goal of the changes in the prompt.\n\n")

		presenter.Step("Gathering environment context...")
		tools.AppendSectionHeader(&outputBuffer, "Environment Context")
		// Pass ctx to helper functions
		osNameOutput, _, osErr := ExecClient.CaptureOutput(ctx, cwd, "uname", "-s")
		if osErr != nil {
			presenter.Warning("Could not determine OS type: %v", osErr)
			fmt.Fprintf(&outputBuffer, "OS Type: Unknown\n")
		} else {
			fmt.Fprintf(&outputBuffer, "OS Type: %s\n", strings.TrimSpace(osNameOutput))
		}
		outputBuffer.WriteString("Key tool versions:\n")
		appendToolVersion(ctx, &outputBuffer, presenter, cwd, "Go", "go", "version")
		appendToolVersion(ctx, &outputBuffer, presenter, cwd, "git", "git", "--version")
		appendToolVersion(ctx, &outputBuffer, presenter, cwd, "gcloud", "gcloud", "version")
		outputBuffer.WriteString("Other potentially relevant tools:\n")
		appendCommandAvailability(ctx, &outputBuffer, presenter, cwd, "jq")
		appendCommandAvailability(ctx, &outputBuffer, presenter, cwd, "tree")
		outputBuffer.WriteString("Relevant environment variables:\n")
		fmt.Fprintf(&outputBuffer, "  GOOGLE_CLOUD_PROJECT: %s\n", os.Getenv("GOOGLE_CLOUD_PROJECT"))
		fmt.Fprintf(&outputBuffer, "  GOOGLE_REGION: %s\n", os.Getenv("GOOGLE_REGION"))
		nixFilePath := filepath.Join(cwd, ".idx", "dev.nix")
		if _, statErr := os.Stat(nixFilePath); statErr == nil {
			outputBuffer.WriteString("Nix environment definition found: .idx/dev.nix\n")
		}
		outputBuffer.WriteString("\n\n")

		presenter.Step("Gathering Git status...")
		tools.AppendSectionHeader(&outputBuffer, "Git Status (Summary)")
		outputBuffer.WriteString("Provides context on recent local changes:\n\n")
		gitStatus, _, statusErr := client.GetStatusShort(ctx)
		if statusErr != nil {
			presenter.Warning("Failed to get git status: %v", statusErr)
			outputBuffer.WriteString("Failed to get git status.\n")
		} else {
			tools.AppendFencedCodeBlock(&outputBuffer, strings.TrimSpace(gitStatus), "")
		}
		outputBuffer.WriteString("\n\n")

		presenter.Step("Gathering project structure...")
		tools.AppendSectionHeader(&outputBuffer, "Project Structure (Top Levels)")
		outputBuffer.WriteString("Directory layout (up to 2 levels deep):\n\n")
		// Pass ctx to ExecClient calls
		treeOutput, _, treeErr := ExecClient.CaptureOutput(ctx, cwd, "tree", "-L", "2", "-a", "-I", treeIgnorePattern)
		structureOutput := ""
		if treeErr != nil {
			presenter.Warning("'tree' command failed or not found, falling back to 'ls'.")
			lsOutput, _, lsErr := ExecClient.CaptureOutput(ctx, cwd, "ls", "-Ap")
			if lsErr != nil {
				presenter.Warning("Fallback 'ls' command failed: %v", lsErr)
				structureOutput = "Could not determine project structure."
			} else {
				structureOutput = strings.TrimSpace(lsOutput)
			}
		} else {
			structureOutput = strings.TrimSpace(treeOutput)
		}
		tools.AppendFencedCodeBlock(&outputBuffer, structureOutput, "")
		outputBuffer.WriteString("\n\n")

		presenter.Step("Listing and filtering project files...")
		gitLsFilesOutput, _, listErr := client.ListTrackedAndCachedFiles(ctx)
		if listErr != nil {
			return listErr
		}
		filesToList := strings.Split(strings.TrimSpace(gitLsFilesOutput), "\n")
		if len(filesToList) == 1 && filesToList[0] == "" {
			filesToList = []string{}
		}
		presenter.Step("Processing %d potential file(s) for inclusion...", len(filesToList))

		tools.AppendSectionHeader(&outputBuffer, "Relevant Code Files Follow")
		includedFiles := make(map[string]bool)
		filesAddedCount := 0

		for _, filePath := range filesToList {
			if filePath == "" {
				continue
			}
			cleanPath := filepath.Clean(filePath)
			isMatch := includeRe.MatchString(cleanPath)
			isExcluded := excludePathsRegex != "" && excludeRe.MatchString(cleanPath)
			var aiExcludedMatch gitignore.Match
			if aiExcluder != nil {
				aiExcludedMatch = aiExcluder.Match(cleanPath)
			}
			shouldExclude := isExcluded || (aiExcludedMatch != nil && aiExcludedMatch.Ignore())
			if !isMatch || shouldExclude {
				continue
			}

			// Pass ctx to appendFileContentToBuffer, though it doesn't use it yet
			err := appendFileContentToBuffer(ctx, &outputBuffer, presenter, cwd, cleanPath, maxSizeBytes)
			if err == nil {
				includedFiles[cleanPath] = true
				filesAddedCount++
			}
		}

		if len(criticalFiles) > 0 {
			presenter.Step("Checking critical files...")
			for _, criticalPath := range criticalFiles {
				cleanCriticalPath := filepath.Clean(criticalPath)
				fullPath := filepath.Join(cwd, cleanCriticalPath)
				shouldExclude := false
				if aiExcluder != nil {
					match := aiExcluder.Match(cleanCriticalPath)
					if match != nil && match.Ignore() {
						shouldExclude = true
					}
				}
				if shouldExclude {
					continue
				}

				if _, statErr := os.Stat(fullPath); statErr == nil {
					if !includedFiles[cleanCriticalPath] {
						presenter.Detail("Including critical file: %s", cleanCriticalPath)
						// Pass ctx to appendFileContentToBuffer
						err := appendFileContentToBuffer(ctx, &outputBuffer, presenter, cwd, cleanCriticalPath, maxSizeBytes)
						if err == nil {
							filesAddedCount++
						}
					}
				} else if !os.IsNotExist(statErr) {
					presenter.Warning("Could not check critical file %s: %v", cleanCriticalPath, statErr)
				}
			}
		}

		presenter.Newline()
		presenter.Step("Writing context file %s (%d files included)...", presenter.Highlight(outputFilePath), filesAddedCount)
		err = tools.WriteBufferToFile(outputFilePath, &outputBuffer) // tools.WriteBufferToFile is fine as it's just file I/O
		if err != nil {
			presenter.Error("Failed to write output file '%s': %v", outputFilePath, err)
			return err
		}
		presenter.Success("Successfully generated context file: %s", outputFilePath)
		return nil
	},
}

// Updated signature to include ctx
func appendToolVersion(ctx context.Context, buf *bytes.Buffer, p *ui.Presenter, cwd, displayName, commandName string, args ...string) {
	_ = ctx // Silences unused parameter warning if ctx is only for ExecClient
	_ = ctx // Silences unused parameter warning if ctx is only for ExecClient
	_ = ctx // Explicitly ignore ctx if only passed through
	fmt.Fprintf(buf, "  %s: ", displayName)
	logger := AppLogger // Assuming AppLogger is accessible as a package variable from cmd/root.go

	// Prefer --version first, use ExecClient
	versionOutput, _, versionErr := ExecClient.CaptureOutput(ctx, cwd, commandName, "--version")
	if versionErr == nil && strings.TrimSpace(versionOutput) != "" {
		output := versionOutput
		parsedOutput := strings.TrimSpace(output)
		if commandName == "go" && strings.HasPrefix(output, "go version") {
			parts := strings.Fields(output)
			if len(parts) >= 3 {
				parsedOutput = parts[2]
			}
		} else if commandName == "git" && strings.HasPrefix(output, "git version") {
			parts := strings.Fields(output)
			if len(parts) >= 3 {
				parsedOutput = parts[2]
			}
		} else if commandName == "gcloud" && strings.Contains(output, "Google Cloud SDK") {
			lines := strings.Split(output, "\n")
			sdkLineFound := false
			for _, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if strings.HasPrefix(trimmedLine, "Google Cloud SDK") {
					parsedOutput = trimmedLine
					sdkLineFound = true
					break
				}
			}
			if !sdkLineFound {
				parsedOutput = strings.SplitN(strings.TrimSpace(output), "\n", 2)[0]
			}
		} else {
			parsedOutput = strings.SplitN(parsedOutput, "\n", 2)[0]
		}
		buf.WriteString(parsedOutput)
		buf.WriteString("\n")
		logger.Debug("Tool version found", slog.String("tool", commandName), slog.String("version", parsedOutput))
		return
	}
	logger.Debug("Tool --version flag failed or gave empty output", slog.String("tool", commandName), slog.Any("error", versionErr))

	// Fallback to original args, use ExecClient
	output, _, err := ExecClient.CaptureOutput(ctx, cwd, commandName, args...)
	if err != nil || strings.TrimSpace(output) == "" {
		buf.WriteString("Not found\n")
		// Use ExecClient.CommandExists
		if !ExecClient.CommandExists(commandName) { // Check with ExecClient
			p.Warning("Required tool '%s' not found in PATH.", commandName)
			logger.Error("Required tool version check failed: not found", slog.String("tool", commandName))
		} else {
			p.Warning("Could not determine version for '%s'.", commandName)
			logger.Warn("Tool version check failed or empty output", slog.String("tool", commandName), slog.Any("error", err))
		}
		return
	}
	// (Parsing logic remains the same as before)
	parsedOutput := strings.TrimSpace(output)
	if commandName == "go" && strings.HasPrefix(output, "go version") {
		parts := strings.Fields(output)
		if len(parts) >= 3 {
			parsedOutput = parts[2]
		}
	} else if commandName == "git" && strings.HasPrefix(output, "git version") {
		parts := strings.Fields(output)
		if len(parts) >= 3 {
			parsedOutput = parts[2]
		}
	} else if commandName == "gcloud" && strings.Contains(output, "Google Cloud SDK") {
		lines := strings.Split(output, "\n")
		sdkLineFound := false
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "Google Cloud SDK") {
				parsedOutput = trimmedLine
				sdkLineFound = true
				break
			}
		}
		if !sdkLineFound {
			parsedOutput = strings.SplitN(strings.TrimSpace(output), "\n", 2)[0]
		}
	} else {
		parsedOutput = strings.SplitN(parsedOutput, "\n", 2)[0]
	}
	buf.WriteString(parsedOutput)
	buf.WriteString("\n")
	logger.Debug("Tool version found (via fallback args)", slog.String("tool", commandName), slog.String("version", parsedOutput))
}

// Updated signature to include ctx (though not used by ExecClient.CommandExists directly)
func appendCommandAvailability(ctx context.Context, buf *bytes.Buffer, p *ui.Presenter, cwd string, commandName string) {
	_ = ctx // Silences unused parameter warning if ctx is only for ExecClient
	_ = ctx // Silences unused parameter warning if ctx is only for ExecClient
	_ = ctx // Explicitly ignore ctx if only passed through
	// Renamed unused parameter from _ to cwd to match the call signature, even if not used directly by CommandExists
	_ = cwd // Explicitly ignore cwd if CommandExists doesn't need it
	fmt.Fprintf(buf, "  %s: ", commandName)
	logger := AppLogger

	// Use ExecClient.CommandExists
	if ExecClient.CommandExists(commandName) {
		buf.WriteString("Available\n")
		logger.Debug("Optional tool available", slog.String("tool", commandName))
	} else {
		buf.WriteString("Not found\n")
		p.Warning("Optional tool '%s' not found in PATH.", commandName)
		logger.Warn("Optional tool check: not found", slog.String("tool", commandName))
	}
}

// Updated signature to include ctx, though not directly used by os.Stat or tools.ReadFileContent
func appendFileContentToBuffer(ctx context.Context, buf *bytes.Buffer, p *ui.Presenter, cwd, filePath string, maxSizeBytes int64) error {
	_ = ctx // Explicitly ignore ctx for now if unused by current logic // Explicitly ignore ctx for now if unused by current logic // Explicitly ignore ctx for now if unused by current logic // Explicitly ignore ctx for now if unused by current logic
	_ = ctx // Explicitly ignore ctx for now if unused by current logic
	_ = ctx // Explicitly ignore ctx for now
	fullPath := filepath.Join(cwd, filePath)
	logger := AppLogger
	logger.Debug("Attempting to append file content", slog.String("path", filePath), slog.String("full_path", fullPath))
	info, err := os.Stat(fullPath)
	if err != nil {
		errMsg := ""
		if os.IsNotExist(err) {
			errMsg = fmt.Sprintf("Skipping '%s' (does not exist)", filePath)
		} else {
			errMsg = fmt.Sprintf("Skipping '%s' (cannot stat: %v)", filePath, err)
		}
		p.Warning(errMsg)
		return errors.New(errMsg)
	}
	if !info.Mode().IsRegular() {
		errMsg := fmt.Sprintf("Skipping '%s' (not a regular file)", filePath)
		p.Warning(errMsg)
		return errors.New(errMsg)
	}
	if info.Size() == 0 {
		logger.Debug("Skipping empty file", slog.String("path", filePath))
		return fmt.Errorf("skipping empty file %s", filePath)
	}
	if info.Size() > maxSizeBytes {
		errMsg := fmt.Sprintf("Skipping '%s' (too large: %dB > %dB limit)", filePath, info.Size(), maxSizeBytes)
		p.Warning(errMsg)
		return errors.New(errMsg)
	}
	content, err := tools.ReadFileContent(fullPath) // tools.ReadFileContent is fine (file I/O)
	if err != nil {
		errMsg := fmt.Sprintf("Skipping '%s' (read error: %v)", filePath, err)
		p.Warning(errMsg)
		return errors.New(errMsg)
	}
	tools.AppendFileMarkerHeader(buf, filePath) // markdown util
	buf.Write(content)
	tools.AppendFileMarkerFooter(buf, filePath) // markdown util
	logger.Debug("Appended file content successfully", slog.String("path", filePath), slog.Int64("size", info.Size()))
	return nil
}

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.Flags().StringVarP(&describeOutputFile, "output", "o", defaultDescribeOutputFile, "Path to write the context markdown file")
}
