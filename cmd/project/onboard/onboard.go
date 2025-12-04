package onboard

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/tools"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/contextvibes/cli/internal/workitem"
	"github.com/contextvibes/cli/internal/workitem/github"
	"github.com/spf13/cobra"
)

//go:embed onboard.md.tpl
var onboardLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var outputFlag string

const (
	defaultSystemPromptPath = ".idx/airules.md"
	maxFileSizeKB           = 500
	concurrentFetches       = 3
	maxTreeDepth            = 2
)

// OnboardCmd represents the project onboard command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var OnboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Generates a complete 'Session Initialization Artifact' for AI onboarding.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Generating AI Session Onboarding Artifact")

		var finalBuffer bytes.Buffer

		// --- Header ---
		fmt.Fprintf(&finalBuffer, "# AI Session Initialization\n")
		fmt.Fprintf(&finalBuffer, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
		//nolint:lll // Long instruction string.
		fmt.Fprintf(&finalBuffer, "> **User Instruction:** I am initializing a new development session. Below is my System Persona (THEA), the current Project Status (Summary), and the Codebase Snapshot (Describe). Ingest this context, acknowledge you are ready, and await my instructions.\n\n")

		// --- Layer 1: System Persona ---
		presenter.Step("Layer 1: Loading System Persona...")
		systemPrompt, err := os.ReadFile(defaultSystemPromptPath)
		if err != nil {
			presenter.Warning("Could not read system prompt at %s: %v", defaultSystemPromptPath, err)
			//nolint:lll // Long line.
			fmt.Fprintf(&finalBuffer, "## 1. System Persona & Rules\n\n(System prompt file not found at %s)\n\n", defaultSystemPromptPath)
		} else {
			fmt.Fprintf(&finalBuffer, "## 1. System Persona & Rules\n\n%s\n\n", string(systemPrompt))
		}

		// --- Layer 2: Strategic Context (Summary) ---
		presenter.Step("Layer 2: Fetching Project Summary...")
		summaryContent, err := generateSummary(ctx)
		if err != nil {
			presenter.Warning("Failed to generate summary: %v", err)
			fmt.Fprintf(&finalBuffer, "## 2. Project Status (Morning Briefing)\n\n(Failed to fetch data)\n\n")
		} else {
			fmt.Fprintf(&finalBuffer, "## 2. Project Status (Morning Briefing)\n\n%s\n\n", summaryContent)
		}

		// --- Layer 3: Technical Context (Describe) ---
		presenter.Step("Layer 3: Snapshotting Codebase...")
		describeContent, err := generateDescribe(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate codebase snapshot: %w", err)
		}
		fmt.Fprintf(&finalBuffer, "## 3. Technical Context\n\n%s\n\n", describeContent)

		// --- Output ---
		err = tools.WriteBufferToFile(outputFlag, &finalBuffer)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		presenter.Success("Onboarding artifact generated: %s", outputFlag)
		presenter.Info("Upload this file to your AI to start the session.")

		return nil
	},
}

// generateSummary fetches issues and formats them as Markdown.
func generateSummary(ctx context.Context) (string, error) {
	provider, err := newProvider(ctx, globals.AppLogger, globals.LoadedAppConfig)
	if err != nil {
		return "", err
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(concurrentFetches)

	var (
		bugs, myTasks, epics        []workitem.WorkItem
		errBugs, errTasks, errEpics error
	)

	go func() {
		defer waitGroup.Done()

		bugs, errBugs = provider.SearchItems(ctx, "is:open is:issue label:bug sort:updated-desc")
	}()

	go func() {
		defer waitGroup.Done()

		myTasks, errTasks = provider.SearchItems(ctx, "is:open is:issue assignee:@me sort:updated-desc")
	}()

	go func() {
		defer waitGroup.Done()

		epics, errEpics = provider.SearchItems(ctx, "is:open is:issue label:epic sort:updated-desc")
	}()

	waitGroup.Wait()

	var buf bytes.Buffer
	formatSection(&buf, "🚨 Urgent Attention (Bugs)", bugs, errBugs)
	formatSection(&buf, "👤 On Your Plate (Assigned)", myTasks, errTasks)
	formatSection(&buf, "🗺️ Strategic Context (Epics)", epics, errEpics)

	return buf.String(), nil
}

func formatSection(buf *bytes.Buffer, title string, items []workitem.WorkItem, err error) {
	buf.WriteString("### " + title + "\n")

	switch {
	case err != nil:
		fmt.Fprintf(buf, "_Error fetching data: %v_\n", err)
	case len(items) == 0:
		buf.WriteString("_None found._\n")
	default:
		for _, item := range items {
			fmt.Fprintf(buf, "- [#%d] %s\n", item.Number, item.Title)
		}
	}

	buf.WriteString("\n")
}

// generateDescribe snapshots the codebase (simplified logic from describe command).
func generateDescribe(ctx context.Context) (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
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
		return "", fmt.Errorf("git init failed: %w", err)
	}

	var buf bytes.Buffer

	// Git Status
	gitStatus, _, statusErr := client.GetStatusShort(ctx)
	if statusErr != nil {
		gitStatus = "Failed to get git status."
	}

	tools.AppendSectionHeader(&buf, "Git Status (Summary)")
	tools.AppendFencedCodeBlock(&buf, strings.TrimSpace(gitStatus), "")

	// Tree (Native Implementation)
	tools.AppendSectionHeader(&buf, "Project Structure")

	treeOutput, err := generateNativeTree(workDir)
	if err != nil {
		treeOutput = "Error generating tree: " + err.Error()
	}

	tools.AppendFencedCodeBlock(&buf, treeOutput, "")

	// Files
	tools.AppendSectionHeader(&buf, "Relevant Code Files")

	gitLsFilesOutput, _, err := client.ListTrackedAndCachedFiles(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list tracked files: %w", err)
	}

	processFiles(&buf, gitLsFilesOutput)

	return buf.String(), nil
}

func processFiles(buf *bytes.Buffer, gitLsFilesOutput string) {
	//nolint:mnd // 1024 is standard KB conversion.
	maxSizeBytes := int64(maxFileSizeKB * 1024)
	filesToList := strings.SplitSeq(strings.TrimSpace(gitLsFilesOutput), "\n")

	for file := range filesToList {
		if file == "" || !shouldIncludeFile(file) {
			continue
		}

		info, statErr := os.Stat(file)
		if statErr != nil || info.Size() > maxSizeBytes {
			continue
		}

		content, readErr := tools.ReadFileContent(file)
		if readErr == nil {
			tools.AppendFileMarkerHeader(buf, file)
			buf.Write(content)
			tools.AppendFileMarkerFooter(buf, file)
		}
	}
}

func shouldIncludeFile(file string) bool {
	// 1. Skip system prompt (already in Layer 1)
	if file == defaultSystemPromptPath {
		return false
	}
	// 2. Skip common noise
	if strings.HasPrefix(file, "vendor/") || strings.HasPrefix(file, ".git/") || strings.HasSuffix(file, "go.sum") {
		return false
	}
	// 3. Skip binary files by extension
	ext := strings.ToLower(filepath.Ext(file))

	return !isBinaryExt(ext)
}

// generateNativeTree walks the directory and produces a tree-like string.
func generateNativeTree(root string) (string, error) {
	var buf bytes.Buffer

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err // Propagate error
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "." {
			return nil
		}

		if isIgnoredDir(entry) {
			return fs.SkipDir
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		if depth > maxTreeDepth {
			if entry.IsDir() {
				return fs.SkipDir
			}

			return nil
		}

		indent := strings.Repeat("  ", depth)

		marker := "|-"
		if depth == 0 {
			marker = ""
		}

		fmt.Fprintf(&buf, "%s%s %s\n", indent, marker, entry.Name())

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	return buf.String(), nil
}

func isIgnoredDir(entry fs.DirEntry) bool {
	if !entry.IsDir() {
		return false
	}

	name := entry.Name()

	return name == ".git" || name == "vendor" || name == "node_modules" || name == ".terraform"
}

func isBinaryExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".exe", ".bin", ".dll", ".so", ".dylib", ".zip", ".tar", ".gz":
		return true
	default:
		return false
	}
}

// newProvider factory (duplicated to avoid circular deps or complex refactor).
//

func newProvider(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
) (workitem.Provider, error) {
	switch cfg.Project.Provider {
	case "github":
		//nolint:wrapcheck // Factory function.
		return github.New(ctx, logger, cfg)
	case "":
		//nolint:wrapcheck // Factory function.
		return github.New(ctx, logger, cfg)
	default:
		//nolint:err113 // Dynamic error is appropriate here.
		return nil, fmt.Errorf("unsupported provider '%s'", cfg.Project.Provider)
	}
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(onboardLongDescription, nil)
	if err != nil {
		panic(err)
	}

	OnboardCmd.Short = desc.Short
	OnboardCmd.Long = desc.Long
	OnboardCmd.Flags().StringVarP(&outputFlag, "output", "o", "_contextvibes.md", "Output file path")
}
