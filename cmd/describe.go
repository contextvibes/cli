// cmd/describe.go

package cmd

import (
	"bytes"
	"context" // Added
	"errors"
	"fmt"
	"log/slog" // Added
	"os"
	"path/filepath"
	"regexp"
	"strings"

	// Use new packages
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"

	// Keep tools for Markdown, file I/O, and *non-git* command execution helpers
	"github.com/contextvibes/cli/internal/tools"

	gitignore "github.com/denormal/go-gitignore"
	"github.com/spf13/cobra"
)

// --- Constants and Critical Files (remain the same) ---
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

// --- End Constants ---

// Assume AppLogger is initialized in rootCmd for AI file logging
// var AppLogger *slog.Logger // Defined in root.go

var describeCmd = &cobra.Command{
	Use:   "describe [-o <output_file>]",
	Short: "Generates project context file (default: contextvibes.md).",
	Long: `Gathers project context (user prompt, environment, git status, structure, relevant files)
and writes it to a Markdown file (default: ` + defaultDescribeOutputFile + `), suitable for AI interaction.

Respects .gitignore, .aiexclude rules, and file size limits when including file content.`,
	Example: `  contextvibes describe                 # Prompts for input, saves context to contextvibes.md
  contextvibes describe -o project_snapshot.md # Saves context to custom file`,
	Args: cobra.NoArgs,
	// Add Silence flags as we handle output/errors via Presenter
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Generating project context description.")

		// --- Init Git Client & Basic Checks ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Describe: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "describe"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init (is this a Git repository?): %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "describe"))
		cwd := client.Path() // Use repo root path from client

		// --- Compile Regexes ---
		logger.DebugContext(ctx, "Compiling regex patterns", slog.String("source_command", "describe"))
		includeRe, err := regexp.Compile(includeExtensionsRegex)
		if err != nil {
			presenter.Error("Internal error: invalid include regex pattern.")
			logger.Error("Invalid include regex", slog.Any("error", err))
			return err
		}
		excludeRe, err := regexp.Compile(excludePathsRegex)
		if err != nil {
			presenter.Error("Internal error: invalid exclude regex pattern.")
			logger.Error("Invalid exclude regex", slog.Any("error", err))
			return err
		}
		maxSizeBytes := int64(maxFileSizeKB * 1024)

		// --- Load .aiexclude patterns ---
		logger.DebugContext(ctx, "Loading .aiexclude patterns", slog.String("source_command", "describe"))
		var aiExcluder gitignore.GitIgnore
		aiExcludeFilePath := filepath.Join(cwd, ".aiexclude")
		// Use os.ReadFile which is simpler
		aiExcludeContent, readErr := os.ReadFile(aiExcludeFilePath)
		gitignoreErrorHandler := func(ignoreErr gitignore.Error) bool {
			presenter.Warning("Parsing .aiexclude file: %v", ignoreErr)
			logger.WarnContext(ctx, "Parsing .aiexclude file", slog.Any("error", ignoreErr))
			return true // Continue parsing other lines
		}
		if readErr == nil {
			aiExcluder = gitignore.New(bytes.NewReader(aiExcludeContent), cwd, gitignoreErrorHandler)
			if aiExcluder != nil {
				presenter.Info("Loaded exclusion rules from %s", presenter.Highlight(".aiexclude"))
				logger.InfoContext(ctx, "Loaded .aiexclude rules", slog.String("source_command", "describe"))
			} else if len(bytes.TrimSpace(aiExcludeContent)) > 0 { // Check if file wasn't just whitespace/comments
				presenter.Info(".aiexclude file found but contains no active rules.") // Be less alarming than Warning
				logger.InfoContext(ctx, ".aiexclude parsed but yielded no rules", slog.String("source_command", "describe"))
			}
		} else if !os.IsNotExist(readErr) {
			presenter.Warning("Could not read .aiexclude file at %s: %v", aiExcludeFilePath, readErr)
			logger.WarnContext(ctx, "Error reading .aiexclude", slog.Any("error", readErr), slog.String("path", aiExcludeFilePath))
		} else {
			logger.DebugContext(ctx, ".aiexclude file not found, skipping.", slog.String("path", aiExcludeFilePath))
		}

		// --- Setup Output ---
		var outputBuffer bytes.Buffer
		outputFilePath := describeOutputFile // From flag
		if outputFilePath == "" {
			outputFilePath = defaultDescribeOutputFile
		} // Default handled by Cobra flag usually
		presenter.Info("Generating context file: %s", presenter.Highlight(outputFilePath))

		// --- Section: User Prompt ---
		tools.AppendSectionHeader(&outputBuffer, "Prompt")
		presenter.Separator()
		presenter.Step("Enter a prompt for the AI (e.g., 'Refactor X module', 'Add Y feature to script').")
		presenter.Step("Be specific about goals, files, resources, or errors.")
		presenter.Separator()
		userPrompt, err := presenter.PromptForInput("> Prompt: ") // Use presenter prompt
		if err != nil {
			logger.ErrorContext(ctx, "Failed reading user prompt", slog.String("error", err.Error()))
			return err
		}
		if userPrompt == "" {
			errMsg := "prompt cannot be empty"
			presenter.Error(errMsg)
			logger.ErrorContext(ctx, "User provided empty prompt")
			return errors.New(errMsg)
		}
		fmt.Fprintf(&outputBuffer, "%s\n\n", userPrompt)
		logger.DebugContext(ctx, "User prompt captured", slog.Int("length", len(userPrompt)))

		// --- Section: Collaboration Notes ---
		tools.AppendSectionHeader(&outputBuffer, "Collaboration Notes")
		outputBuffer.WriteString("For future reviews:\n")
		outputBuffer.WriteString("- If code changes are significant or span multiple areas, please provide the full updated file(s) using this task.\n")
		outputBuffer.WriteString("- If changes are small and localized (e.g., fixing a typo, a few lines in one function), you can provide just the relevant snippet, but clearly state the filename and function/context.\n")
		outputBuffer.WriteString("- Always describe the goal of the changes in the prompt.\n\n")

		// --- Section: Environment Context ---
		presenter.Step("Gathering environment context...") // Use Step for progress
		logger.DebugContext(ctx, "Gathering environment context", slog.String("source_command", "describe"))
		tools.AppendSectionHeader(&outputBuffer, "Environment Context")
		osNameOutput, _, osErr := tools.CaptureCommandOutput(cwd, "uname", "-s") // Keep using tools helper for non-git commands for now
		if osErr != nil {
			presenter.Warning("Could not determine OS type: %v", osErr)
			logger.WarnContext(ctx, "uname -s failed", slog.Any("error", osErr))
			fmt.Fprintf(&outputBuffer, "OS Type: Unknown\n")
		} else {
			fmt.Fprintf(&outputBuffer, "OS Type: %s\n", strings.TrimSpace(osNameOutput))
		}
		outputBuffer.WriteString("Key tool versions:\n")
		// Call modified helpers, passing presenter
		appendToolVersion(&outputBuffer, presenter, cwd, "Go", "go", "version")
		appendToolVersion(&outputBuffer, presenter, cwd, "git", "git", "--version")
		appendToolVersion(&outputBuffer, presenter, cwd, "gcloud", "gcloud", "version")
		outputBuffer.WriteString("Other potentially relevant tools:\n")
		appendCommandAvailability(&outputBuffer, presenter, cwd, "jq")
		appendCommandAvailability(&outputBuffer, presenter, cwd, "tree")
		outputBuffer.WriteString("Relevant environment variables:\n")
		fmt.Fprintf(&outputBuffer, "  GOOGLE_CLOUD_PROJECT: %s\n", os.Getenv("GOOGLE_CLOUD_PROJECT"))
		fmt.Fprintf(&outputBuffer, "  GOOGLE_REGION: %s\n", os.Getenv("GOOGLE_REGION"))
		nixFilePath := filepath.Join(cwd, ".idx", "dev.nix")
		if _, statErr := os.Stat(nixFilePath); statErr == nil {
			outputBuffer.WriteString("Nix environment definition found: .idx/dev.nix\n")
			logger.DebugContext(ctx, "Nix file found", slog.String("path", nixFilePath))
		} else if !os.IsNotExist(statErr) {
			logger.WarnContext(ctx, "Failed to stat nix file", slog.String("path", nixFilePath), slog.Any("error", statErr))
		}
		outputBuffer.WriteString("\n\n")

		// --- Section: Git Status (Summary) ---
		presenter.Step("Gathering Git status...") // Use Step for progress
		logger.DebugContext(ctx, "Getting git status --short", slog.String("source_command", "describe"))
		tools.AppendSectionHeader(&outputBuffer, "Git Status (Summary)")
		outputBuffer.WriteString("Provides context on recent local changes:\n\n")
		gitStatus, _, statusErr := client.GetStatusShort(ctx) // Use GitClient method
		if statusErr != nil {
			presenter.Warning("Failed to get git status: %v", statusErr)
			logger.WarnContext(ctx, "Failed to get git status", slog.String("source_command", "describe"), slog.Any("error", statusErr))
			outputBuffer.WriteString("Failed to get git status.\n")
		} else {
			tools.AppendFencedCodeBlock(&outputBuffer, strings.TrimSpace(gitStatus), "")
		}
		outputBuffer.WriteString("\n\n")

		// --- Section: Project Structure (Top Levels) ---
		presenter.Step("Gathering project structure...") // Use Step for progress
		logger.DebugContext(ctx, "Getting project structure", slog.String("source_command", "describe"))
		tools.AppendSectionHeader(&outputBuffer, "Project Structure (Top Levels)")
		outputBuffer.WriteString("Directory layout (up to 2 levels deep):\n\n")
		treeOutput, _, treeErr := tools.CaptureCommandOutput(cwd, "tree", "-L", "2", "-a", "-I", treeIgnorePattern) // Keep using tools helper
		structureOutput := ""
		if treeErr != nil {
			presenter.Warning("'tree' command failed or not found, falling back to 'ls'.")
			logger.WarnContext(ctx, "'tree' command failed", slog.Any("error", treeErr))
			lsOutput, _, lsErr := tools.CaptureCommandOutput(cwd, "ls", "-Ap")
			if lsErr != nil {
				presenter.Warning("Fallback 'ls' command failed: %v", lsErr)
				logger.WarnContext(ctx, "'ls' command failed", slog.Any("error", lsErr))
				structureOutput = "Could not determine project structure."
			} else {
				structureOutput = strings.TrimSpace(lsOutput)
			}
		} else {
			structureOutput = strings.TrimSpace(treeOutput)
		}
		tools.AppendFencedCodeBlock(&outputBuffer, structureOutput, "")
		outputBuffer.WriteString("\n\n")

		// --- Section: Relevant File Contents ---
		presenter.Step("Listing and filtering project files...") // Use Step for progress
		logger.DebugContext(ctx, "Listing files via GitClient", slog.String("source_command", "describe"))
		// Use GitClient method to list files
		gitLsFilesOutput, _, listErr := client.ListTrackedAndCachedFiles(ctx) // Use Correct Method Name
		if listErr != nil {
			presenter.Error("Failed to list files using Git: %v", listErr)
			return listErr
		}

		filesToList := strings.Split(strings.TrimSpace(gitLsFilesOutput), "\n")
		if len(filesToList) == 1 && filesToList[0] == "" { // Handle empty output case
			filesToList = []string{}
		}
		presenter.Step("Processing %d potential file(s) for inclusion...", len(filesToList)) // Use Step for progress
		logger.DebugContext(ctx, "Processing file list", slog.Int("potential_files", len(filesToList)))

		tools.AppendSectionHeader(&outputBuffer, "Relevant Code Files Follow")
		includedFiles := make(map[string]bool)
		filesAddedCount := 0
		filesSkippedCount := 0

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

			logger.Debug("File filter check", slog.String("path", cleanPath), slog.Bool("match", isMatch), slog.Bool("excluded", isExcluded), slog.Bool("aiExcluded", aiExcludedMatch != nil && aiExcludedMatch.Ignore()))

			if !isMatch || shouldExclude {
				filesSkippedCount++
				continue
			}

			// Append content (use modified helper)
			err := appendFileContentToBuffer(&outputBuffer, presenter, cwd, cleanPath, maxSizeBytes)
			if err != nil {
				logger.Warn("Skipping file content append", slog.String("path", cleanPath), slog.Any("reason", err))
				filesSkippedCount++
			} else {
				includedFiles[cleanPath] = true
				filesAddedCount++
			}
		}
		logger.DebugContext(ctx, "Finished processing initial file list.", slog.Int("included_count", filesAddedCount), slog.Int("skipped_count", filesSkippedCount))

		// --- Explicit Critical File Additions ---
		if len(criticalFiles) > 0 {
			presenter.Step("Checking critical files...") // Use Step for progress
			logger.DebugContext(ctx, "Checking critical files", slog.String("source_command", "describe"))
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
					logger.Debug("Skipping critical file due to .aiexclude", slog.String("path", cleanCriticalPath))
					continue
				}

				if _, statErr := os.Stat(fullPath); statErr == nil {
					if !includedFiles[cleanCriticalPath] {
						// Use Detail for less prominent user output here
						presenter.Detail("Including critical file: %s", cleanCriticalPath)
						logger.Info("Explicitly adding critical file", slog.String("path", cleanCriticalPath))
						err := appendFileContentToBuffer(&outputBuffer, presenter, cwd, cleanCriticalPath, maxSizeBytes)
						if err != nil {
							logger.Warn("Skipping critical file content append", slog.String("path", cleanCriticalPath), slog.Any("reason", err))
						} else {
							filesAddedCount++
						}
					} else {
						logger.Debug("Critical file already included", slog.String("path", cleanCriticalPath))
					}
				} else if !os.IsNotExist(statErr) {
					presenter.Warning("Could not check critical file %s: %v", cleanCriticalPath, statErr)
					logger.Warn("Failed to stat critical file", slog.String("path", cleanCriticalPath), slog.Any("error", statErr))
				}
			}
		}

		// --- Write Buffer to File ---
		presenter.Newline()
		presenter.Step("Writing context file %s (%d files included)...", presenter.Highlight(outputFilePath), filesAddedCount) // Use Step
		err = tools.WriteBufferToFile(outputFilePath, &outputBuffer)                                                           // Keep using tools helper for now
		if err != nil {
			presenter.Error("Failed to write output file '%s': %v", outputFilePath, err)
			logger.Error("Failed writing describe output", slog.String("path", outputFilePath), slog.Any("error", err))
			return err
		}
		// Make WriteBufferToFile silent later? For now, let it print its INFO.
		presenter.Success("Successfully generated context file: %s", outputFilePath)
		logger.Info("Describe context file generated successfully", slog.String("path", outputFilePath), slog.Int("files_included", filesAddedCount))
		return nil
	},
}

// --- Modified Internal Helper Functions ---
// They now accept *ui.Presenter and use it for warnings/errors

func appendToolVersion(buf *bytes.Buffer, p *ui.Presenter, cwd, displayName, commandName string, args ...string) {
	fmt.Fprintf(buf, "  %s: ", displayName) // Write to markdown buffer
	logger := AppLogger

	// Prefer --version first
	versionOutput, _, versionErr := tools.CaptureCommandOutput(cwd, commandName, "--version")
	if versionErr == nil && strings.TrimSpace(versionOutput) != "" {
		output := versionOutput
		parsedOutput := strings.TrimSpace(output)
		// ... (parsing logic) ...
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

	// Fallback to original args
	output, _, err := tools.CaptureCommandOutput(cwd, commandName, args...)
	if err != nil || strings.TrimSpace(output) == "" {
		buf.WriteString("Not found\n")
		if !tools.CommandExists(commandName) {
			p.Warning("Required tool '%s' not found in PATH.", commandName) // Use Presenter Warning
			logger.Error("Required tool version check failed: not found", slog.String("tool", commandName))
		} else {
			p.Warning("Could not determine version for '%s'.", commandName) // Use Presenter Warning
			logger.Warn("Tool version check failed or empty output", slog.String("tool", commandName), slog.Any("error", err))
		}
		return
	}

	parsedOutput := strings.TrimSpace(output)
	// ... (parsing logic) ...
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

func appendCommandAvailability(buf *bytes.Buffer, p *ui.Presenter, _, commandName string) {
	fmt.Fprintf(buf, "  %s: ", commandName)
	logger := AppLogger
	if tools.CommandExists(commandName) {
		buf.WriteString("Available\n")
		logger.Debug("Optional tool available", slog.String("tool", commandName))
	} else {
		buf.WriteString("Not found\n")
		p.Warning("Optional tool '%s' not found in PATH.", commandName) // Use Presenter Warning
		logger.Warn("Optional tool check: not found", slog.String("tool", commandName))
	}
}

func appendFileContentToBuffer(buf *bytes.Buffer, p *ui.Presenter, cwd, filePath string, maxSizeBytes int64) error {
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
	content, err := tools.ReadFileContent(fullPath)
	if err != nil {
		errMsg := fmt.Sprintf("Skipping '%s' (read error: %v)", filePath, err)
		p.Warning(errMsg)
		return errors.New(errMsg)
	}
	tools.AppendFileMarkerHeader(buf, filePath)
	buf.Write(content)
	tools.AppendFileMarkerFooter(buf, filePath)
	logger.Debug("Appended file content successfully", slog.String("path", filePath), slog.Int64("size", info.Size()))
	return nil
}

func init() {
	rootCmd.AddCommand(describeCmd)
	describeCmd.Flags().StringVarP(&describeOutputFile, "output", "o", defaultDescribeOutputFile, "Path to write the context markdown file")
}
