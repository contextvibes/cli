// cmd/project/describe/describe.go
package describe

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	gitignore "github.com/denormal/go-gitignore"
	"github.com/spf13/cobra"
)

//go:embed describe.md.tpl
var describeLongDescription string

var describeOutputFile string

const (
	defaultDescribeOutputFile = "contextvibes.md"
	includeExtensionsRegex    = `\.(go|mod|sum|tf|py|yaml|json|md|gitignore|txt|hcl|nix)$|^(Taskfile\.yaml|requirements\.txt|README\.md|\.idx/dev\.nix|\.idx/airules\.md)$`
	maxFileSizeKB             = 500
	excludePathsRegex         = `(^vendor/|^\.git/|^\.terraform/|^\.venv/|^__pycache__/|^\.DS_Store|^\.pytest_cache/|^\.vscode/|\.tfstate|\.tfplan|^secrets?/|\.auto\.tfvars|ai_context\.txt|crash.*\.log|contextvibes\.md)`
	treeIgnorePattern         = "vendor|.git|.terraform|.venv|venv|env|__pycache__|.pytest_cache|.DS_Store|.idx|.vscode|*.tfstate*|*.log|ai_context.txt|contextvibes.md|node_modules|build|dist"
)

// DescribeCmd represents the describe command
var DescribeCmd = &cobra.Command{
	Use:     "describe [-o <output_file>]",
	Example: `  contextvibes project describe -o project_snapshot.md`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		
		logger, ok := cmd.Context().Value("logger").(*slog.Logger)
		if !ok { return errors.New("logger not found in context") }
		execClient, ok := cmd.Context().Value("execClient").(*exec.ExecutorClient)
		if !ok { return errors.New("execClient not found in context") }
		loadedAppConfig, ok := cmd.Context().Value("config").(*config.Config)
		if !ok { return errors.New("config not found in context") }
		ctx := cmd.Context()

		presenter.Summary("Generating project context description.")

		workDir, err := os.Getwd()
		if err != nil { return err }

		gitCfg := git.GitClientConfig{
			Logger:                logger,
			DefaultRemoteName:     loadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: loadedAppConfig.Git.DefaultMainBranch,
			Executor:              execClient.UnderlyingExecutor(),
		}
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		cwd := client.Path()

		includeRe, _ := regexp.Compile(includeExtensionsRegex)
		excludeRe, _ := regexp.Compile(excludePathsRegex)
		maxSizeBytes := int64(maxFileSizeKB * 1024)

		var aiExcluder gitignore.GitIgnore
		aiExcludeFilePath := filepath.Join(cwd, ".aiexclude")
		aiExcludeContent, readErr := os.ReadFile(aiExcludeFilePath)
		if readErr == nil {
			aiExcluder = gitignore.New(bytes.NewReader(aiExcludeContent), cwd, nil)
		}

		var outputBuffer bytes.Buffer
		userPrompt, err := presenter.PromptForInput("Enter a prompt for the AI: ")
		if err != nil || userPrompt == "" {
			return errors.New("prompt cannot be empty")
		}
		fmt.Fprintf(&outputBuffer, "### Prompt\n\n%s\n\n", userPrompt)

		gitStatus, _, statusErr := client.GetStatusShort(ctx)
		if statusErr != nil {
			gitStatus = "Failed to get git status."
		}
		tools.AppendSectionHeader(&outputBuffer, "Git Status (Summary)")
		tools.AppendFencedCodeBlock(&outputBuffer, strings.TrimSpace(gitStatus), "")

		treeOutput, _, treeErr := execClient.CaptureOutput(ctx, workDir, "tree", "-L", "2", "-a", "-I", treeIgnorePattern)
		if treeErr != nil {
			treeOutput = "Could not generate tree view. 'tree' command may not be installed."
		}
		tools.AppendSectionHeader(&outputBuffer, "Project Structure")
		tools.AppendFencedCodeBlock(&outputBuffer, strings.TrimSpace(treeOutput), "")

		tools.AppendSectionHeader(&outputBuffer, "Relevant Code Files")
		
		gitLsFilesOutput, _, err := client.ListTrackedAndCachedFiles(ctx)
		if err != nil {
			return err
		}
		filesToList := strings.Split(strings.TrimSpace(gitLsFilesOutput), "\n")

		for _, file := range filesToList {
			if file == "" { continue }
			
			if !includeRe.MatchString(file) || excludeRe.MatchString(file) {
				continue
			}
			if aiExcluder != nil && aiExcluder.Match(file) != nil && aiExcluder.Match(file).Ignore() {
				continue
			}
			
			info, statErr := os.Stat(file)
			if statErr != nil { continue }
			if info.Size() > maxSizeBytes { continue }

			content, readErr := tools.ReadFileContent(file)
			if readErr == nil {
				tools.AppendFileMarkerHeader(&outputBuffer, file)
				outputBuffer.Write(content)
				tools.AppendFileMarkerFooter(&outputBuffer, file)
			}
		}

		if err := tools.WriteBufferToFile(describeOutputFile, &outputBuffer); err != nil {
			return err
		}

		presenter.Success("Successfully generated context file: %s", describeOutputFile)
		return nil
	},
}

func init() {
	desc, err := cmddocs.ParseAndExecute(describeLongDescription, nil)
	if err != nil {
		panic(err)
	}
	DescribeCmd.Short = desc.Short
	DescribeCmd.Long = desc.Long
	DescribeCmd.Flags().StringVarP(&describeOutputFile, "output", "o", "contextvibes.md", "Path to write the context markdown file")
}
