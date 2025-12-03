// Package describe provides the command to generate a project context description.
package describe

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	gitignore "github.com/denormal/go-gitignore"
	"github.com/spf13/cobra"
)

//go:embed describe.md.tpl
var describeLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	describeOutputFile string
	describePromptFlag string
)

const (
	maxFileSizeKB = 500
	//nolint:lll // Pattern is long.
	treeIgnorePattern = "vendor|.git|.terraform|.venv|venv|env|__pycache__|.pytest_cache|.DS_Store|.idx|.vscode|*.tfstate*|*.log|ai_context.txt|contextvibes.md|node_modules|build|dist"
)

// DescribeCmd represents the describe command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var DescribeCmd = &cobra.Command{
	Use:     "describe [-o <output_file>]",
	Example: `  contextvibes project describe -o project_snapshot.md`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Generating project context description.")

		workDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		//nolint:exhaustruct // Partial config is sufficient.
		gitCfg := git.GitClientConfig{
			Logger:                globals.AppLogger,
			DefaultRemoteName:     globals.LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: globals.LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              globals.ExecClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)

			return fmt.Errorf("failed to initialize git client: %w", err)
		}
		cwd := client.Path()

		// Compile include/exclude patterns from config
		includeRes := make(
			[]*regexp.Regexp,
			0,
			len(globals.LoadedAppConfig.Describe.IncludePatterns),
		)
		for _, p := range globals.LoadedAppConfig.Describe.IncludePatterns {
			re, err := regexp.Compile(p)
			if err != nil {
				return fmt.Errorf("invalid include pattern in config '%s': %w", p, err)
			}
			includeRes = append(includeRes, re)
		}

		excludeRes := make(
			[]*regexp.Regexp,
			0,
			len(globals.LoadedAppConfig.Describe.ExcludePatterns),
		)
		for _, p := range globals.LoadedAppConfig.Describe.ExcludePatterns {
			re, err := regexp.Compile(p)
			if err != nil {
				return fmt.Errorf("invalid exclude pattern in config '%s': %w", p, err)
			}
			excludeRes = append(excludeRes, re)
		}

		//nolint:mnd // 1024 is standard KB conversion.
		maxSizeBytes := int64(maxFileSizeKB * 1024)

		var aiExcluder gitignore.GitIgnore
		aiExcludeFilePath := filepath.Join(cwd, ".aiexclude")
		//nolint:gosec // Reading .aiexclude is intended.
		aiExcludeContent, readErr := os.ReadFile(aiExcludeFilePath)
		if readErr == nil {
			aiExcluder = gitignore.New(bytes.NewReader(aiExcludeContent), cwd, nil)
		}

		var outputBuffer bytes.Buffer
		var userPrompt string
		if describePromptFlag != "" {
			userPrompt = describePromptFlag
		} else {
			var promptErr error
			userPrompt, promptErr = presenter.PromptForInput("Enter a prompt for the AI: ")
			if promptErr != nil {
				return fmt.Errorf("prompt input failed: %w", promptErr)
			}
		}

		if userPrompt == "" {
			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("prompt cannot be empty")
		}

		fmt.Fprintf(&outputBuffer, "### Prompt\n\n%s\n\n", userPrompt)

		gitStatus, _, statusErr := client.GetStatusShort(ctx)
		if statusErr != nil {
			gitStatus = "Failed to get git status."
		}
		tools.AppendSectionHeader(&outputBuffer, "Git Status (Summary)")
		tools.AppendFencedCodeBlock(&outputBuffer, strings.TrimSpace(gitStatus), "")

		treeOutput, _, treeErr := globals.ExecClient.CaptureOutput(
			ctx,
			workDir,
			"tree",
			"-L",
			"2",
			"-a",
			"-I",
			treeIgnorePattern,
		)
		if treeErr != nil {
			treeOutput = "Could not generate tree view. 'tree' command may not be installed."
		}
		tools.AppendSectionHeader(&outputBuffer, "Project Structure")
		tools.AppendFencedCodeBlock(&outputBuffer, strings.TrimSpace(treeOutput), "")

		tools.AppendSectionHeader(&outputBuffer, "Relevant Code Files")

		gitLsFilesOutput, _, err := client.ListTrackedAndCachedFiles(ctx)
		if err != nil {
			return fmt.Errorf("failed to list git files: %w", err)
		}
		filesToList := strings.SplitSeq(strings.TrimSpace(gitLsFilesOutput), "\n")

		for file := range filesToList {
			if file == "" {
				continue
			}

			// New configurable matching logic
			isIncluded := false
			for _, re := range includeRes {
				if re.MatchString(file) {
					isIncluded = true

					break
				}
			}
			if !isIncluded {
				continue
			}

			isExcluded := false
			for _, re := range excludeRes {
				if re.MatchString(file) {
					isExcluded = true

					break
				}
			}
			if isExcluded {
				continue
			}

			if aiExcluder != nil && aiExcluder.Match(file) != nil &&
				aiExcluder.Match(file).Ignore() {
				continue
			}

			info, statErr := os.Stat(file)
			if statErr != nil {
				continue
			}
			if info.Size() > maxSizeBytes {
				continue
			}

			content, readErr := tools.ReadFileContent(file)
			if readErr == nil {
				tools.AppendFileMarkerHeader(&outputBuffer, file)
				outputBuffer.Write(content)
				tools.AppendFileMarkerFooter(&outputBuffer, file)
			}
		}

		//nolint:noinlineerr // Inline check is standard.
		if err := tools.WriteBufferToFile(describeOutputFile, &outputBuffer); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		presenter.Success("Successfully generated context file: %s", describeOutputFile)

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(describeLongDescription, nil)
	if err != nil {
		panic(err)
	}

	DescribeCmd.Short = desc.Short
	DescribeCmd.Long = desc.Long
	DescribeCmd.Flags().
		StringVarP(&describeOutputFile, "output", "o", "contextvibes.md", "Path to write the context markdown file")
	// THE FIX: This line defines the flag so Cobra knows about it.
	DescribeCmd.Flags().
		StringVarP(&describePromptFlag, "prompt", "p", "", "Provide the prompt text directly")
}
