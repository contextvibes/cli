### Prompt

Project: ContextVibes CLI (github.com/contextvibes/cli)

### Collaboration Notes

For future reviews:
- If code changes are significant or span multiple areas, please provide the full updated file(s) using this task.
- If changes are small and localized (e.g., fixing a typo, a few lines in one function), you can provide just the relevant snippet, but clearly state the filename and function/context.
- Always describe the goal of the changes in the prompt.

### Environment Context

OS Type: Linux
Key tool versions:
  Go: go1.24.2
  git: 2.47.2
  gcloud: Google Cloud SDK 494.0.0
Other potentially relevant tools:
  jq: Available
  tree: Not found
Relevant environment variables:
  GOOGLE_CLOUD_PROJECT: 
  GOOGLE_REGION: 
Nix environment definition found: .idx/dev.nix


### Git Status (Summary)

Provides context on recent local changes:

```
M cmd/kickoff.go
 M internal/config/config.go
 M internal/exec/client.go
?? .contextvibes.yaml
?? DEVELOPMENT.md
?? internal/kickoff/
?? kickoff.md
?? my_kickoff_prompt_for_describe.txt
?? notes.txt
```



### Project Structure (Top Levels)

Directory layout (up to 2 levels deep):

```
aiexclude
bin/
CHANGELOG.md
cmd/
codemod.json
contextvibes
contextvibes_ai_trace.log
contextvibes.log
contextvibes.md
.contextvibes.yaml
CONTRIBUTING.md
DEVELOPMENT.md
docs/
.git/
.gitignore
go.mod
go.sum
.idx/
internal/
kickoff.md
LICENSE
my_kickoff_prompt_for_describe.txt
notes.txt
README.md
ROADMAP.md
```



### Relevant Code Files Follow

======== FILE: .contextvibes.yaml ========

# .contextvibes.yaml
# Sample configuration file for the ContextVibes CLI.
# This file allows you to customize default behaviors.
# Settings here will be merged with built-in defaults.
# Command-line flags override settings in this file.

# --- Git Workflow Settings ---
git:
  # defaultRemote: "origin" # Default is 'origin'
  # defaultMainBranch: "main" # Default is 'main'

# --- Logging Settings ---
logging:
  # defaultAILogFile: "contextvibes_ai_trace.log" # Default if not set
  # defaultAILogFile: "logs/ai_activity.jsonl" # Example: custom path

# --- Validation Rules ---
validation:
  branchName:
    # enable: true # Default is true. Set to false to disable validation.
    # pattern: "^((feature|fix|docs|format)/.+)$" # Default pattern
    # Example custom pattern: enforce JIRA-style prefixes for features/fixes
    # pattern: "^((feature|fix)/(JIRA-[0-9]+)-.+|(docs|chore)/.+)$"

  commitMessage:
    # enable: true # Default is true. Set to false to disable validation.
    # pattern: "^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+$" # Default Conventional Commits pattern
    # Example custom pattern:
    # pattern: "^(TASK-[0-9]+|MERGE): .+"

# --- Project State (Managed by ContextVibes CLI) ---
# This section might be automatically updated by certain ContextVibes commands.
projectState:
  # strategicKickoffCompleted: false # Default is false. Set to true by 'contextvibes kickoff --strategic' upon successful completion.
  # lastStrategicKickoffDate: "YYYY-MM-DDTHH:MM:SSZ" # Optional: timestamp of last strategic kickoff

# --- AI Collaboration Preferences (Managed by ContextVibes CLI) ---
# This section can be populated during a strategic kickoff session to store
# user preferences for how the AI assistant (ContextVibes) should behave.
ai:
  collaborationPreferences:
    # codeProvisioningStyle: "bash_cat_eof" # Options: "bash_cat_eof", "raw_markdown"
    # markdownDocsStyle: "raw_markdown" # Typically "raw_markdown"
    # detailedTaskMode: "mode_b" # Options: "mode_a" (Generate & Refine), "mode_b" (Interactive Step-by-Step)
    # proactiveDetailLevel: "detailed_explanations" # Options: "detailed_explanations", "concise_unless_asked"
    # aiProactivity: "proactive_suggestions" # Options: "proactive_suggestions", "wait_for_request"
    # Note: These keys are illustrative; actual keys to be defined during implementation.

======== END FILE: .contextvibes.yaml ========

======== FILE: DEVELOPMENT.md ========

# Local Development Guide for ContextVibes CLI

This guide provides instructions for setting up your local development environment to work on the `contextvibes` Go CLI, run tests, and follow project conventions.

## Prerequisites

Before you begin, ensure you have the following installed and configured:

1.  **Git:** For version control. [Installation Guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).
2.  **Go:** Refer to the `go.mod` file for the specific version (e.g., 1.24.x or later). It's recommended to manage Go versions using a tool like `gvm` or asdf, or ensure your system Go matches. [Official Go Downloads](https://go.dev/dl/).
3.  **(Optional but Recommended) Nix:** If a `.idx/dev.nix` or `flake.nix` is present in the future for this project, using Nix can help create a reproducible development environment with all tools.
4.  **External Tools for Full Command Testing:** To test all `contextvibes` commands, you'll need the tools it wraps installed and in your PATH:
    *   Terraform CLI (`terraform`, `tflint`)
    *   Pulumi CLI (`pulumi`)
    *   Python (`python`, `pip`), and Python tools (`isort`, `black`, `flake8`)
    *   Other Go tools if used by quality checks (e.g., `golangci-lint`)

## Initial Setup

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/contextvibes/cli.git
    cd cli
    ```

2.  **(If using Nix) Enter the Nix Development Environment:**
    ```bash
    # Example if a flake.nix is added later:
    # nix develop .#
    ```
    If a Nix environment is defined, it will make specific versions of Go and other tools available.

3.  **Install Go Module Dependencies:**
    ```bash
    go mod download
    go mod tidy
    ```

4.  **Build the CLI:**
    You can build the CLI for your local system:
    ```bash
    go build -o contextvibes ./cmd/contextvibes/main.go
    ```
    You can then run it as `./contextvibes`. Alternatively, install it to your $GOPATH/bin:
    ```bash
    go install ./cmd/contextvibes/main.go
    # Ensure /bin is in your PATH
    ```

## Common Development Tasks

This project uses standard Go commands. You can also use a development build of `contextvibes` itself to manage its own workflow (dogfooding).

### 1. Running Unit Tests

Run all unit tests (excluding integration tests, if any are tagged separately):
```bash
go test ./...
```
To detect race conditions (highly recommended):
```bash
go test -race ./...
```
To get coverage information:
```bash
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### 2. Running Integration Tests (If Applicable)

If integration tests exist (e.g., in an `integration_test/` directory with build tags):
```bash
# Example: go test -tags=integration -v ./integration_test/...
# Refer to specific integration testing docs if available.
```

### 3. Linting

This project may use `golangci-lint` (check for a `.golangci.yml` file). If so:
```bash
golangci-lint run ./...
```
Also, always run:
```bash
go vet ./...
```
You can use a development build of `contextvibes quality` on its own codebase too.

### 4. Formatting Code

Ensure your code is formatted according to Go standards:
```bash
go fmt ./...
# If goimports-reviser or similar is standard for the project:
# goimports-reviser -format ./...
```
Alternatively, use `./contextvibes format` (your dev build).

### 5. Tidying Go Modules
After adding or updating dependencies:
```bash
go mod tidy
```
Commit both `go.mod` and `go.sum` after changes.

## Debugging

The Go ecosystem includes `delve` for debugging.
You can run tests with Delve or attach it to a running process.
For VS Code users, the Go extension provides debugging capabilities that should work with Delve.

Example of running a specific test with Delve:
```bash
# dlv test ./path/to/package -test.run TestSpecificFunction
dlv test ./internal/config -test.run TestConfigLoading # Example
```

## Using `contextvibes` for its Own Development

Once you have a working build of `contextvibes`, you are encouraged to use it for managing your development workflow on the CLI itself:

*   **Daily Branches:** `./contextvibes kickoff -b feature/my-new-cli-feature`
*   **Committing:** `./contextvibes commit -m "feat(command): Add new flag"`
    *   This will use the commit message validation rules defined in `.contextvibes.yaml` (once created for this repo).
*   **Formatting/Quality:** `./contextvibes format` and `./contextvibes quality`
*   **Syncing:** `./contextvibes sync`

## Configuration for Development (`.contextvibes.yaml`)

It's recommended to have a `.contextvibes.yaml` file in the root of this CLI's repository for testing its own configuration loading features. You can start with the sample provided in the documentation or generate one if an `init-config` command is added.

---

This guide should help you get started with developing the ContextVibes CLI.

======== END FILE: DEVELOPMENT.md ========

======== FILE: internal/kickoff/assets/strategic_kickoff_protocol_template.md ========

# Master Prompt for Strategic Project Kickoff: {{.ProjectName}}

## AI Facilitator Instructions & Persona (For You, The AI Assistant Reading This Prompt)

You are an **Expert AI Project Kickoff Facilitator**. Your primary objective is to guide the user (a software developer or technical lead) meticulously and strategically through a project kickoff for their software project named **'{{.ProjectName}}'**, which they've described as a **'{{.ProjectAppType}}'** application.

**Your Core Tasks:**
1.  Follow the structured **"Project Kickoff Protocol"** detailed below (starting from Module A).
2.  Ask the user questions from each section and module of the protocol, one by one.
3.  When specific project data is needed by you to better understand their context (e.g., current Git status, file content, project structure for architectural discussions), **instruct the user to run the appropriate ` + "`contextvibes`" + ` CLI command** in their terminal (e.g., ` + "`contextvibes describe -o project_context.md`" + `, ` + "`contextvibes status`" + `, ` + "`contextvibes diff`" + `) and to **paste the relevant output back into this chat for your analysis.**
4.  For certain sections (like AI Collaboration Preferences for how *ContextVibes CLI itself* should behave, or Git settings for `.contextvibes.yaml`), if the protocol below specifies, synthesize the user's choices into the **structured YAML format** provided in the protocol. Present this YAML clearly to the user and instruct them to save it into their project's `.contextvibes.yaml` file, usually under a specified key.
5.  Maintain a helpful, guiding, and expert persona. Use clear, professional language.
6.  At the end of the protocol (Module E), instruct the user to create a comprehensive summary document of all decisions made during this session (e.g., ` + "`PROJECT_KICKOFF_SUMMARY.md`" + `).
7.  Finally, instruct the user to inform the ` + "`contextvibes`" + ` CLI that this strategic kickoff is complete by running a command like ` + "`contextvibes kickoff --mark-strategic-complete`" + ` in their terminal for project '{{.ProjectName}}'.

**Key ` + "`contextvibes`" + ` commands you might ask the user to run during your interaction:**
*   ` + "`contextvibes describe -o project_context.md`" + `: For comprehensive project context (environment, git status, structure, key files).
*   ` + "`contextvibes status`" + `: For current Git status summary.
*   ` + "`contextvibes diff`" + `: For pending Git changes.
*   ` + "`contextvibes test`" + `: To understand current test status or discuss testing strategy.
*   ` + "`contextvibes quality`" + `: For code quality insights.
(The user is expected to know how to run these ` + "`contextvibes`" + ` commands if they are using the tool).

### Your AI Facilitator Interaction Style with the User:
*   **Detailed Task Interaction Model:** Use **Mode B (Interactive Step-by-Step Co-Creation)**. Present one item or concept from the checklist at a time; wait for the user's response, then proceed.
*   **Depth of Explanation & Proactivity:** Provide **detailed explanations** for concepts if they seem complex. Be **proactive in suggesting relevant software engineering patterns, libraries, or potential challenges** based on the user's project type and responses. Reference best practices.
*   **Clarity:** Ask clear and unambiguous questions.

---
## Project Kickoff Protocol for '{{.ProjectName}}' (Type: '{{.ProjectAppType}}')
(You, the AI Facilitator, will now guide the user through these modules and items sequentially)

**Module A: Project Definition & Scope Clarification (AI to guide user)**

Let's start by deeply understanding your project's definition and scope.

  **A.1: Reconfirm Core Problem & Solution:**
    *   The user initially described the project '{{.ProjectName}}' as solving: ` + "`[User's Core Idea - To be asked or confirmed if known]`" + `.
    *   AI to User: "Based on any initial thoughts for project '{{.ProjectName}}', could you confirm or refine its core problem and the proposed solution? What primary challenge does it address?"

  **A.2: Articulating the Project Vision and ContextVibes' Strategic Role (if user plans to use ContextVibes for *this* project):**
    *   AI to User: "What is the broader vision for '{{.ProjectName}}'? If you plan to use the ContextVibes CLI during the development of '{{.ProjectName}}', how do you see its AI-augmented capabilities (like context generation with ` + "`describe`" + `, or workflow automation) contributing to this project's specific strategic goals?"

  **A.3: Target Audience/Users:**
    *   AI to User: "Who are the primary users or consumers of '{{.ProjectName}}'? (e.g., end-users, other services, developers using it as a library)."

  **A.4: Measurable Outcomes for Project & Tool Integration:**
    *   AI to User: "What would make '{{.ProjectName}}' a clear success? Let's think about 1-2 measurable metrics. If you're using AI tools (including ContextVibes to feed an AI) for its development, are there specific outcomes you expect from that AI usage (e.g., time saved, quality improvement)?"

  **A.5: Review Constraints, Dependencies, Assumptions, Risks & Non-Goals:**
    *   AI to User: "Let's consolidate any known technical constraints (platforms, languages), resource limitations, critical external dependencies (including any AI services you'll rely on), key assumptions, initial high-level risks (technical, project, AI-related), or explicit non-goals for '{{.ProjectName}}'."
    *   AI to User: "To help identify these, you can run ` + "`contextvibes describe -o initial_context.md`" + ` in your project directory and share relevant sections with me, especially regarding dependencies or environment context."

**Module B: ContextVibes Tooling (Your CLI) & Technical Deep Dive for '{{.ProjectName}}'**

Now let's discuss the technical foundations for '{{.ProjectName}}', and how ContextVibes CLI might be used in its development.

  **B.1: Relevant ContextVibes Commands & Features for *This* Project:**
    *   AI to User: "Which specific ContextVibes commands (` + "`describe`" + `, ` + "`diff`" + `, ` + "`kickoff`" + `, ` + "`commit`" + `, ` + "`format`" + `, ` + "`quality`" + `, ` + "`test`" + `, ` + "`codemod`" + `, etc.) do you anticipate using most for developing '{{.ProjectName}}'? How do you envision them helping?"

  **--- AI Facilitator Instruction: Generate YAML for *ContextVibes CLI's* AI Collaboration Preferences ---**
  After discussing collaboration preferences with the user for how *they want ContextVibes CLI's own AI features to behave in the future if it were to directly call AI for them (this is hypothetical for now as ContextVibes CLI currently generates context for the user to take to an AI)*, OR how they prefer *you (this current AI assistant)* to provide information during *this kickoff session* (these preferences should have been captured by the ContextVibes CLI itself before generating this master prompt), please present the following YAML snippet to the user. This YAML represents the preferences that the ContextVibes CLI noted. Instruct the user to ensure this is saved in their project's `.contextvibes.yaml` file under ` + "`ai.collaborationPreferences`" + ` so that ` + "`contextvibes kickoff`" + ` (and potentially other future AI features in the CLI) can use these preferences.

  AI to User: "ContextVibes CLI recorded the following interaction preferences when we started. Please ensure this YAML block is saved in your project's `.contextvibes.yaml` under the ` + "`ai.collaborationPreferences`" + ` key:"

{{.CollaborationPrefsYAML}}
{{/* This will be replaced by the Go code with the actual YAML string */}}

  **--- End AI Facilitator Instruction ---**

  **B.2 - B.9: Technical Foundations for '{{.ProjectName}}' (AI to guide user through these from the protocol, asking user to run ` + "`contextvibes describe`" + ` or specific commands as needed for context)**
    *   B.2: Scope of Tool Usage (ContextVibes or other tools for *this* project)
    *   B.3: Technical Architecture & ContextVibes Integration (if any beyond context prep)
    *   B.4: AI Provider/Model Strategy for *this* project (if it uses AI directly)
    *   B.5: Module/Project Setup (` + "`go.mod`" + `, ` + "`requirements.txt`" + ` etc.)
    *   B.6: Key External Dependencies
    *   B.7: Configuration Management Strategy
    *   B.8: Authentication/Authorization Strategy
    *   B.9: Data Management & Persistence Strategy

**Module C: Execution Framework & Development Practices for '{{.ProjectName}}' (AI to guide user)**
{{/* You will need to paste the full content of Module C from the protocol here, ensuring placeholders like {{.ProjectName}} are used where appropriate. */}}
{{/* Example for one item: */}}
  **C.1: Roles & Responsibilities:**
    *   AI to User: "Let's define the key roles and responsibilities for the '{{.ProjectName}}' project team."
{{/* ... continue with all items C.1 through C.10 ... */}}
  **C.3c: ContextVibes Usage Conventions for developing '{{.ProjectName}}' (Prompt engineering for AI, review of AI outputs, etc.).**

**Module D: Governance, Communication, and Risk Management for '{{.ProjectName}}' (AI to guide user)**
{{/* You will need to paste the full content of Module D from the protocol here. */}}
{{/* Example for one item: */}}
  **D.2: Risk Management:**
    *   AI to User: "What are the key risks for '{{.ProjectName}}'? Consider technical risks, project risks, and any risks related to using AI tools."

**Module E: Conclusion and Immediate Next Steps for '{{.ProjectName}}' (AI to guide user)**
{{/* You will need to paste the full content of Module E from the protocol here. */}}
{{/* Example for one item: */}}
  **E.2: Summarize Action Items:**
    *   AI to User: "Let's summarize the key Action Items, their Owners, and target Deadlines that came out of our discussion for '{{.ProjectName}}'."
  **E.6: Final Instruction:**
    *   AI to User: "Excellent. To complete this process, please create a summary document of our key decisions (e.g., in ` + "`PROJECT_KICKOFF_SUMMARY.md`" + `). Then, in your terminal for project '{{.ProjectName}}', run ` + "`contextvibes kickoff --mark-strategic-complete`" + ` (or an equivalent command ContextVibes provides) to let the CLI know this strategic kickoff is done. This will enable the daily Git workflow mode for future ` + "`contextvibes kickoff`" + ` runs in this project."

---
End of Master Kickoff Prompt for AI Facilitator.

======== END FILE: internal/kickoff/assets/strategic_kickoff_protocol_template.md ========

======== FILE: internal/kickoff/orchestrator.go ========

// Package kickoff provides the orchestration logic for the ContextVibes CLI's
// 'kickoff' command, handling both strategic project kickoffs and daily
// Git workflow startups.
package kickoff

import (
	"bytes" // For text/template
	"context"
	_ "embed" // Required for //go:embed
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template" // Required for template processing
	"time"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
)

var assumeYes bool

const strategicKickoffPromptFilename = "STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md"

//go:embed assets/strategic_kickoff_protocol_template.md
var masterKickoffProtocolTemplateContent string // Content will be embedded here

type Orchestrator struct {
	logger         *slog.Logger
	config         *config.Config
	presenter      *ui.Presenter
	gitClient      *git.GitClient
	configFilePath string
}

func NewOrchestrator(
	logger *slog.Logger,
	cfg *config.Config,
	presenter *ui.Presenter,
	gitClient *git.GitClient,
	configFilePath string,
) *Orchestrator {
	if logger == nil {
		fmt.Fprintln(os.Stderr, "[WARN] Kickoff Orchestrator initialized with a nil logger. Using discard logger.")
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Orchestrator{
		logger:         logger.With("component", "kickoff.Orchestrator"),
		config:         cfg,
		presenter:      presenter,
		gitClient:      gitClient,
		configFilePath: configFilePath,
	}
}

func (o *Orchestrator) ExecuteKickoff(ctx context.Context, isStrategicFlag bool, branchNameFlag string, globalAssumeYes bool) error {
	assumeYes = globalAssumeYes
	o.logger.DebugContext(ctx, "ExecuteKickoff called",
		slog.Bool("isStrategicFlag", isStrategicFlag),
		slog.String("branchNameFlag", branchNameFlag),
		slog.Bool("assumeYesGlobal", assumeYes))

	if o.config == nil {
		err := errors.New("orchestrator config is nil, cannot proceed")
		o.logger.ErrorContext(ctx, "Configuration error in ExecuteKickoff", slog.Any("error", err))
		o.presenter.Error("Internal error: Kickoff orchestrator missing essential configuration.")
		return err
	}

	runStrategic := isStrategicFlag
	if !runStrategic {
		if o.config.ProjectState.StrategicKickoffCompleted == nil || !*o.config.ProjectState.StrategicKickoffCompleted {
			o.presenter.Info("No prior strategic kickoff detected for this project.")
			o.presenter.Info("This command will now guide you to generate a master prompt for an AI to facilitate a full Strategic Project Kickoff.")
			o.presenter.Newline()
			runStrategic = true
		}
	}

	if runStrategic {
		return o.executeStrategicKickoffGeneration(ctx)
	}
	return o.executeDailyKickoff(ctx, branchNameFlag)
}

func (o *Orchestrator) MarkStrategicKickoffComplete(ctx context.Context) error {
	o.presenter.Summary("Marking Strategic Project Kickoff as Complete...")
	o.logger.InfoContext(ctx, "Attempting to mark strategic kickoff as complete.")

	if o.config == nil {
		err := errors.New("orchestrator config is nil, cannot mark complete")
		o.logger.ErrorContext(ctx, "Configuration error in MarkStrategicKickoffComplete", slog.Any("error", err))
		o.presenter.Error("Internal error: Configuration not loaded.")
		return err
	}

	trueVal := true
	o.config.ProjectState.StrategicKickoffCompleted = &trueVal
	o.config.ProjectState.LastStrategicKickoffDate = time.Now().UTC().Format(time.RFC3339)

	if err := config.UpdateAndSaveConfig(o.config, o.configFilePath); err != nil {
		errMsg := fmt.Sprintf("failed to save updated configuration to %s", o.configFilePath)
		o.presenter.Error("%s: %v", errMsg, err)
		o.logger.ErrorContext(ctx, "Failed to save config for marking strategic kickoff complete", "error", err, "path", o.configFilePath)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	o.presenter.Success("Strategic kickoff has been marked as complete in: %s", o.presenter.Highlight(o.configFilePath))
	o.presenter.Info("Subsequent `contextvibes kickoff` runs (without --strategic) will now perform the daily Git workflow.")
	o.logger.InfoContext(ctx, "Strategic kickoff marked as complete.", "path", o.configFilePath)
	return nil
}

func (o *Orchestrator) executeStrategicKickoffGeneration(ctx context.Context) error {
	o.presenter.Summary("ContextVibes: Strategic Project Kickoff - Prompt Generation")
	o.logger.InfoContext(ctx, "Starting Strategic Project Kickoff prompt generation.")

	if err := o.runCollaborationSetup(ctx); err != nil {
		return fmt.Errorf("phase IV (Collaboration Setup for CLI) failed: %w", err)
	}
	initialProjectInfo, errGathering := o.runInitialInfoGathering(ctx)
	if errGathering != nil {
		return fmt.Errorf("phase I (Initial Info Gathering for CLI) failed: %w", errGathering)
	}
	if err := o.runTechnicalReadinessInquiry(ctx); err != nil {
		return fmt.Errorf("phase II (Technical Readiness Inquiry for CLI) failed: %w", err)
	}

	o.presenter.Newline()
	o.presenter.Step("Generating the Master Kickoff Prompt for your AI assistant...")

	promptText, err := o.generateMasterKickoffPromptText(initialProjectInfo)
	if err != nil {
		// Error already logged by generateMasterKickoffPromptText
		o.presenter.Error("Failed to generate the master kickoff prompt content: %v", err)
		return err
	}

	promptFilePath := filepath.Join(".", strategicKickoffPromptFilename)
	err = os.WriteFile(promptFilePath, []byte(promptText), 0644)
	if err != nil {
		o.presenter.Error("Failed to save the master kickoff prompt to '%s': %v", promptFilePath, err)
		o.logger.ErrorContext(ctx, "Failed to write master kickoff prompt file", "path", promptFilePath, "error", err)
		return fmt.Errorf("failed to save master kickoff prompt: %w", err)
	}
	o.presenter.Success("Master Kickoff Prompt successfully generated and saved to: %s", o.presenter.Highlight(promptFilePath))
	o.logger.InfoContext(ctx, "Master kickoff prompt generated and saved.", "path", promptFilePath)

	o.presenter.Newline()
	o.presenter.Header("Next Steps for Your Strategic Kickoff:")
	o.presenter.Info("1. Open the generated file: %s", o.presenter.Highlight(promptFilePath))
	o.presenter.Info("2. Copy its entire content.")
	o.presenter.Info("3. Paste it as the initial prompt to your preferred AI assistant (e.g., Gemini, Claude, ChatGPT).")
	o.presenter.Info("4. The AI will then guide you through the detailed strategic kickoff process.")
	o.presenter.Advice("   During your session with the AI, it may ask you to run specific `contextvibes` commands")
	o.presenter.Advice("   (like `describe` or `status`) in your terminal. Provide the output back to the AI.")
	o.presenter.Newline()
	o.presenter.Info("After completing the session with your AI and creating a project summary:")
	o.presenter.Advice(" - Consider saving any AI-generated YAML configurations to your '.contextvibes.yaml' or other project files.")
	o.presenter.Advice(" - You can then inform ContextVibes by running: `contextvibes kickoff --mark-strategic-complete`.")

	o.logger.InfoContext(ctx, "Strategic Project Kickoff prompt generation finished successfully.")
	return nil
}

func (o *Orchestrator) runCollaborationSetup(ctx context.Context) error {
	o.logger.DebugContext(ctx, "Executing Phase IV: Collaboration Model Setup (for ContextVibes CLI interaction)")
	o.presenter.Header("IV. ContextVibes CLI Interaction Preferences")
	o.presenter.Info("Let's set preferences for how I, ContextVibes CLI, will provide outputs like code or docs if generated based on your AI session summary later.")
	o.presenter.Detail("Your choices can be stored in '%s' if you complete a strategic kickoff and this CLI is enhanced to save them.", o.configFilePath)
	o.presenter.Newline()

	if o.config == nil { return errors.New("config is nil in runCollaborationSetup") }
	prefs := &o.config.AI.CollaborationPreferences

	askAndSetPreference := func(promptKey, currentValue string, validOptions []string, exampleValue string, updateFunc func(string)) error {
		var currentDisplay string
		if currentValue == "" { currentDisplay = "(default or not set)"
		} else { currentDisplay = fmt.Sprintf("current: %s", currentValue) }
		optionsStr := strings.Join(validOptions, " | ")
		promptMsgStr := fmt.Sprintf("%s\n  Options: [ %s ] (e.g., '%s')\n  (%s)\nYour choice (or press Enter to keep current/default): ",
			promptKey, optionsStr, exampleValue, currentDisplay)

		userInput, err := o.presenter.PromptForInput(promptMsgStr)
		if err != nil { return fmt.Errorf("failed to get user input for '%s': %w", promptKey, err) }
		userInput = strings.TrimSpace(userInput)

		if userInput == "" {
			o.presenter.Info("  -> No change for '%s'. Using: '%s'", promptKey, currentValue)
			o.logger.DebugContext(ctx, "Collaboration preference unchanged by user", "key", promptKey, "value", currentValue)
			return nil
		}
		isValidOption := false
		for _, opt := range validOptions { if userInput == opt { isValidOption = true; break } }
		if !isValidOption {
			o.presenter.Warning("  -> Invalid option '%s' for '%s'. Kept: '%s'.", userInput, promptKey, currentValue)
			o.logger.WarnContext(ctx, "Invalid collaboration preference option entered", "key", promptKey, "entered_value", userInput, "kept_value", currentValue)
			return nil 
		}
		updateFunc(userInput)
		o.presenter.Info("  -> Preference for '%s' set to: '%s'", promptKey, userInput)
		o.logger.InfoContext(ctx, "Collaboration preference updated", "key", promptKey, "new_value", userInput)
		return nil
	}
	
	var err error
	err = askAndSetPreference("1. CLI Code/Command Provisioning Style:", prefs.CodeProvisioningStyle, []string{"bash_cat_eof", "raw_markdown"}, "bash_cat_eof", func(s string) { prefs.CodeProvisioningStyle = s })
	if err != nil { return err }
	err = askAndSetPreference("2. CLI Markdown Doc Presentation Style:", prefs.MarkdownDocsStyle, []string{"raw_markdown"}, "raw_markdown", func(s string) { prefs.MarkdownDocsStyle = s })
	if err != nil { return err }
	
	o.presenter.Newline()
	o.presenter.Success("ContextVibes CLI interaction preferences noted (will be saved if strategic kickoff is marked complete).")
	o.logger.InfoContext(ctx, "ContextVibes CLI collaboration model setup phase complete.", slog.Any("final_preferences_in_config_struct", *prefs))
	return nil
}

func (o *Orchestrator) runInitialInfoGathering(ctx context.Context) (map[string]string, error) {
	o.logger.DebugContext(ctx, "Executing Simplified Initial Info Gathering for Prompt Parameterization")
	o.presenter.Header("Project Details for Kickoff Prompt Customization")
	o.presenter.Info("I need a few details to customize the master kickoff prompt for your AI.")
	o.presenter.Newline()

	gatheredInfo := make(map[string]string)
	var err error
	var tempStr string

	tempStr, err = o.presenter.PromptForInput("  What is the official name for this project? (e.g., MyNewService): ")
	if err != nil { return nil, err }
	if tempStr == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil { 
			o.logger.ErrorContext(ctx, "Failed to get current working directory for default project name", slog.Any("error", cwdErr))
			return nil, fmt.Errorf("failed to get current directory for default project name: %w", cwdErr)
		}
		gatheredInfo["projectName"] = filepath.Base(cwd)
	} else {
		gatheredInfo["projectName"] = tempStr
	}
	
	tempStr, err = o.presenter.PromptForInput(fmt.Sprintf("  What is the primary application type being developed for '%s'? (e.g., Go API, Python CLI): ", gatheredInfo["projectName"]))
	if err != nil { return nil, err }
	if tempStr == "" { return nil, errors.New("project application type cannot be empty") }
	gatheredInfo["projectAppType"] = tempStr

	tempStr, err = o.presenter.PromptForInput(fmt.Sprintf("  Is '%s' a brand new project, or an existing one entering a new phase? [new/existing]: ", gatheredInfo["projectName"]))
	if err != nil { return nil, err }
	if tempStr == "" { return nil, errors.New("project stage (new/existing) cannot be empty") }
	gatheredInfo["projectStage"] = tempStr
	
	o.presenter.Newline()
	o.logger.InfoContext(ctx, "Simplified initial info gathered.", slog.Any("info", gatheredInfo))
	return gatheredInfo, nil
}

func (o *Orchestrator) runTechnicalReadinessInquiry(ctx context.Context) error {
	o.logger.DebugContext(ctx, "Executing Phase II: Technical Readiness Inquiry (for ContextVibes CLI user)")
	o.presenter.Header("II. ContextVibes CLI Readiness Check")
	o.presenter.Info("Just a couple of quick checks regarding your `contextvibes` CLI environment:")
	o.presenter.Newline()

	var err error
	_, err = o.presenter.PromptForConfirmation("  1. Is the `contextvibes` CLI installed and accessible in your system PATH? [Y/n]: ")
	if err != nil { return err }

	_, err = o.presenter.PromptForConfirmation("  2. Are any environment variables required by `contextvibes` for its general operation set correctly? [Y/n]: ")
	if err != nil { return err }
	o.presenter.Newline()
	
	o.logger.InfoContext(ctx, "Technical readiness inquiry phase (for CLI user) finished.")
	return nil
}

// generateCollaborationPrefsYAML is a helper to create the YAML string for AI collaboration preferences.
func (o *Orchestrator) generateCollaborationPrefsYAML() string {
    // Ensure defaults if some preferences are empty, to make the YAML more complete for the AI prompt.
    prefs := o.config.AI.CollaborationPreferences
    
    style := prefs.CodeProvisioningStyle
    if style == "" { style = "bash_cat_eof" } // Example default
    
    mdStyle := prefs.MarkdownDocsStyle
    if mdStyle == "" { mdStyle = "raw_markdown" }

    taskMode := prefs.DetailedTaskMode
    if taskMode == "" { taskMode = "mode_b" }

    detailLevel := prefs.ProactiveDetailLevel
    if detailLevel == "" && taskMode == "mode_b" { detailLevel = "detailed_explanations" }
	 if detailLevel == "" { detailLevel = "concise_unless_asked" }


    proactivity := prefs.AIProactivity
    if proactivity == "" { proactivity = "proactive_suggestions" }

    var sb strings.Builder
    sb.WriteString("ai:\n")
    sb.WriteString("  collaborationPreferences:\n")
    sb.WriteString(fmt.Sprintf("    codeProvisioningStyle: \"%s\"\n", style))
    sb.WriteString(fmt.Sprintf("    markdownDocsStyle: \"%s\"\n", mdStyle))
    sb.WriteString(fmt.Sprintf("    detailedTaskMode: \"%s\"\n", taskMode))
    sb.WriteString(fmt.Sprintf("    proactiveDetailLevel: \"%s\"\n", detailLevel))
    sb.WriteString(fmt.Sprintf("    aiProactivity: \"%s\"\n", proactivity))
    return sb.String()
}

func (o *Orchestrator) generateMasterKickoffPromptText(initialInfo map[string]string) (string, error) {
	projectName := initialInfo["projectName"]
	if projectName == "" { projectName = "[User to Specify Project Name]" }
	projectAppType := initialInfo["projectAppType"]
	if projectAppType == "" { projectAppType = "[User to Specify Project Type]" }

	collaborationPrefsYAML := o.generateCollaborationPrefsYAML()

	templateData := struct {
		ProjectName            string
		ProjectAppType         string
		CollaborationPrefsYAML string
	}{
		ProjectName:            projectName,
		ProjectAppType:         projectAppType,
		CollaborationPrefsYAML: collaborationPrefsYAML,
	}

	// Ensure masterKickoffProtocolTemplateContent is not empty
	if strings.TrimSpace(masterKickoffProtocolTemplateContent) == "" {
		errMsg := "embedded master kickoff protocol template is empty"
		o.logger.ErrorContext(context.Background(), errMsg) // Use a background context if ctx not available
		return "", errors.New(errMsg)
	}

	tmpl, err := template.New("kickoffProtocol").Parse(masterKickoffProtocolTemplateContent)
	if err != nil {
		o.logger.Error("Failed to parse master kickoff protocol template", slog.Any("error", err))
		return "", fmt.Errorf("failed to parse master kickoff protocol template: %w", err)
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.Execute(&outputBuffer, templateData); err != nil {
		o.logger.Error("Failed to execute master kickoff protocol template", slog.Any("error", err))
		return "", fmt.Errorf("failed to execute master kickoff protocol template: %w", err)
	}
	
	return outputBuffer.String(), nil
}

func (o *Orchestrator) executeDailyKickoff(ctx context.Context, branchNameFlag string) error {
	o.presenter.Summary("Starting Daily Git Kickoff workflow.")
	o.logger.InfoContext(ctx, "Starting Daily Git Kickoff workflow.")

	if o.gitClient == nil {
		errMsg := "git client not available for daily kickoff. Ensure you are in a Git repository" 
		o.presenter.Error(errMsg)
		o.logger.ErrorContext(ctx, errMsg)
		return errors.New(errMsg)
	}

	mainBranchName := o.config.Git.DefaultMainBranch
	if mainBranchName == "" { mainBranchName = config.DefaultGitMainBranch } 
	remoteName := o.config.Git.DefaultRemote
	if remoteName == "" { remoteName = config.DefaultGitRemote } 
	
	o.presenter.Info("Using remote '%s' and main branch '%s'.", remoteName, mainBranchName)
	o.presenter.Info("Checking Git prerequisites for daily kickoff...")

	currentBranch, err := o.gitClient.GetCurrentBranchName(ctx)
	if err != nil { o.presenter.Error("Failed to get current Git branch: %v", err); return err }
	if currentBranch != mainBranchName {
		errMsg := fmt.Sprintf("not on the main branch ('%s'). Current branch: '%s'", mainBranchName, currentBranch)
		o.presenter.Error(errMsg)
		o.presenter.Advice("Switch to the main branch first using `git switch %s`.", mainBranchName)
		return errors.New(errMsg)
	}
	o.presenter.Info("Confirmed on main branch '%s'.", mainBranchName)

	isClean, err := o.gitClient.IsWorkingDirClean(ctx)
	if err != nil { o.presenter.Error("Failed checking working directory status: %v", err); return err }
	if !isClean {
		errMsg := "working directory is not clean. Daily kickoff requires a clean state on main"
		o.presenter.Error(errMsg)
		o.presenter.Advice("Commit or stash changes first. Try `contextvibes commit -m \"...\"` or `git stash`.")
		return errors.New(errMsg)
	}
	o.presenter.Info("Working directory is clean.")

	targetBranchName := strings.TrimSpace(branchNameFlag)
	branchValidationRule := o.config.Validation.BranchName
	validationIsEnabled := branchValidationRule.Enable == nil || *branchValidationRule.Enable
	effectivePattern := branchValidationRule.Pattern
	
	if validationIsEnabled && effectivePattern == "" {
		effectivePattern = config.DefaultBranchNamePattern 
		o.logger.DebugContext(ctx, "Branch name validation enabled, using default pattern due to empty config pattern", "pattern", effectivePattern)
	}

	if targetBranchName == "" {
		if assumeYes { 
			errMsg := "branch name is required via --branch flag when using --yes for daily kickoff"
			o.presenter.Error(errMsg)
			o.logger.ErrorContext(ctx, errMsg)
			return errors.New(errMsg)
		}
		o.presenter.Newline()
		o.presenter.Info("Please provide the name for the new daily/feature branch.")
		if validationIsEnabled {
			o.presenter.Advice("Pattern to match: %s", effectivePattern)
		}
		for {
			targetBranchName, err = o.presenter.PromptForInput("New branch name: ")
			if err != nil { return err }
			targetBranchName = strings.TrimSpace(targetBranchName)
			if targetBranchName != "" { break }
			o.presenter.Warning("Branch name cannot be empty.")
		}
	}

	if validationIsEnabled {
		branchNameRe, errRe := regexp.Compile(effectivePattern)
		if errRe != nil {
			o.presenter.Error("Internal error: Invalid branch name validation pattern ('%s') in configuration: %v", effectivePattern, errRe)
			return fmt.Errorf("invalid branch name regex in config: %w", errRe)
		}
		if !branchNameRe.MatchString(targetBranchName) {
			errMsg := fmt.Sprintf("invalid branch name: '%s'", targetBranchName) 
			o.presenter.Error(errMsg)
			o.presenter.Advice("Branch name must match the configured pattern: %s", effectivePattern)
			return errors.New(errMsg)
		}
		o.logger.DebugContext(ctx, "Branch name format validated successfully", "pattern", effectivePattern, "branch", targetBranchName)
	} else {
		o.logger.InfoContext(ctx, "Branch name validation is disabled by configuration.")
	}
	
	existsLocally, err := o.gitClient.LocalBranchExists(ctx, targetBranchName)
	if err != nil { o.presenter.Error("Failed checking if branch '%s' exists locally: %v", targetBranchName, err); return err }
	if existsLocally {
		o.presenter.Error("Branch '%s' already exists locally.", targetBranchName)
		return errors.New("branch already exists locally for daily kickoff")
	}
	
	o.presenter.Newline()
	o.presenter.Info("Proposed Daily Kickoff Actions:")
	o.presenter.Detail("1. Update main branch '%s' from remote '%s' (git pull --rebase).", mainBranchName, remoteName)
	o.presenter.Detail("2. Create and switch to new local branch '%s' from '%s'.", targetBranchName, mainBranchName)
	o.presenter.Detail("3. Push new branch '%s' to '%s' and set upstream tracking.", targetBranchName, remoteName)
	o.presenter.Newline()

	if !assumeYes {
		confirmed, promptErr := o.presenter.PromptForConfirmation("Proceed with Daily Kickoff Git workflow?")
		if promptErr != nil { return promptErr }
		if !confirmed {
			o.presenter.Info("Daily kickoff aborted by user.")
			return nil
		}
	} else {
		o.presenter.Info("Confirmation for Daily Kickoff Git workflow bypassed via --yes flag.")
	}

	o.presenter.Newline()
	o.presenter.Step("Updating main branch '%s' from '%s'...", mainBranchName, remoteName)
	if err := o.gitClient.PullRebase(ctx, mainBranchName); err != nil { o.presenter.Error("Failed to update main branch: %v", err); return err }
	o.presenter.Info("Main branch update successful.")

	o.presenter.Newline()
	o.presenter.Step("Creating and switching to new branch '%s' from '%s'...", targetBranchName, mainBranchName)
	if err := o.gitClient.CreateAndSwitchBranch(ctx, targetBranchName, mainBranchName); err != nil { o.presenter.Error("Failed to create/switch branch: %v", err); return err }
	o.presenter.Info("Successfully on new branch '%s'.", targetBranchName)

	o.presenter.Newline()
	o.presenter.Step("Pushing new branch '%s' to '%s' and setting upstream...", targetBranchName, remoteName)
	if err := o.gitClient.PushAndSetUpstream(ctx, targetBranchName); err != nil { o.presenter.Error("Failed to push new branch: %v", err); return err }
	o.presenter.Info("New branch '%s' pushed and upstream tracking set.", targetBranchName)
	
	o.presenter.Newline()
	o.presenter.Success("Daily Git Kickoff workflow completed. You are now on branch '%s'.", targetBranchName)
	o.logger.InfoContext(ctx, "Daily Git Kickoff workflow finished successfully.", slog.String("new_branch", targetBranchName))
	return nil
}

======== END FILE: internal/kickoff/orchestrator.go ========

======== FILE: kickoff.md ========

Okay, "ok" to confirm the action items, responsibilities, and the understanding for Items E.2, E.3, E.4, E.5, and E.6.

We've now completed all modules of the **Strategic Kickoff for the `contextvibes` CLI project**, with a specific focus on planning the integration of the "Enhanced Project Kickoff Protocol" into the `contextvibes kickoff` command.

This was a very thorough session! I will now provide the structured summary of this kickoff, as per Item E.1 of the protocol. This summary will encapsulate all our key decisions and the plan for `contextvibes` CLI development, specifically for this new feature.

---

### Project Kickoff Summary: `contextvibes` CLI - Enhanced Kickoff Feature

**Date of Kickoff:** 2025-05-10
**Project Focus:** Strategic planning for the `contextvibes` CLI project itself, with a primary goal of integrating the "Enhanced Project Kickoff Protocol" into the `contextvibes kickoff` command.
**Project Owner:** You (the user)
**AI Facilitator:** Gemini (as ContextVibes CLI's internal AI engine)

**I. Collaboration Model & Preferences (User-AI for `contextvibes` Development):**

*   **Code Provisioning:** AI to provide Go code for `contextvibes` wrapped in `bash cat <<EOF ... EOF` blocks.
*   **Markdown Docs:** AI to provide `contextvibes` documentation as raw Markdown.
*   **Interaction Model:** Mode B (Interactive Step-by-Step Co-Creation); AI to proactively offer detailed explanations.
*   **Explanation Depth:** AI to provide detailed explanations, referencing Go best practices.
*   **Context Refresh:** User accepts potential need for context refresh in long chat sessions.
*   **Feedback:** Ongoing feedback welcome; end-of-kickoff review confirmed.
*   **`airules.md`:** AI will use general software/Go CLI best practices, adapting from the provided `airules.md` (which was for a Cloud Run API).

**II. Initial Information & ContextVibes Environment Understanding (for `contextvibes` CLI):**

*   **Project Name:** ContextVibes CLI
*   **Module Path:** `github.com/contextvibes/cli`
*   **Go Version (Assumed):** `1.24.x` (to be confirmed from its `go.mod`).
*   **Core Idea (Refined):** Address developer workflow fragmentation and cognitive overhead, especially in preparing project context for AI-assisted development, by providing a unified, intelligent CLI that standardizes common tasks, automates AI context generation, and facilitates programmatic code changes, thereby enhancing individual developer productivity and the quality of AI-assisted software engineering.
*   **Project State:** Existing project (v0.0.5/v0.0.6) entering a new development phase for the enhanced kickoff feature.
*   **Project Type:** Go CLI Application using Cobra. Structure developed organically.
*   **ContextVibes Tool & AI Interaction (Current):**
    *   The Go Cobra CLI (`contextvibes`) currently **does not directly call external AI provider APIs.**
    *   It facilitates AI interaction by generating context (`contextvibes.md`) for the user, who then uses an external AI tool.
    *   The `mcp-vibe-tools` (Python server) represents a potential future architecture or related tool, not current direct capability of this Go CLI.
    *   Thus, `contextvibes` itself doesn't manage AI provider API keys for direct calls yet.
*   **Documentation:** Existing `README.md`, `CHANGELOG.md`, `ROADMAP.md`, etc., are available. AI will suggest improvements as we work.
*   **Primary Goal (This Kickoff):** Strategically plan the integration of the "Enhanced Project Kickoff Protocol" into the `contextvibes kickoff` command. Related roadmap items can be discussed if they align and don't derail.
*   **Development Setup:**
    *   Tested from source.
    *   Action Item: Create a developer-oriented `.contextvibes.yaml` for testing the CLI's own config handling.
    *   Dev environment needs external tools (Git, TF, Python tools, etc.) for full command testing.

**III. Pre-Kickoff Technical Readiness (for `contextvibes` CLI development):**

*   Generally in good shape. Key action is to create a sample/test `.contextvibes.yaml` for the CLI's own development and testing.

**IV. Project Kickoff Checklist - Key Decisions for `contextvibes` CLI & Enhanced Kickoff Feature:**

**Module A: Project Definition & Scope Clarification (for `contextvibes` CLI)**
*   **A.1 Core Problem & Solution:** Refined to emphasize addressing workflow fragmentation and enhancing AI-assisted software engineering through context preparation and standardized tooling.
*   **A.2 Project Vision & Strategic Role:** Vision is to be a key component of an Internal Developer Platform (IDP), boosting **individual developer productivity**. The enhanced kickoff feature aligns by standardizing project initiation and embedding AI-assisted strategic planning.
*   **A.3 Target Audience:** Diverse developers (Go, Python, Terraform, etc.) within an IDP ecosystem; varying experience levels. Pain points: context switching, CLI syntax variety, Git complexity, AI context prep.
*   **A.4 Measurable Outcomes:** Primary: **High Developer Satisfaction** (qualitative feedback, potential scores). Secondary: adoption, workflow consistency, feature delivery (including this kickoff feature).
*   **A.5 Constraints, Dependencies, Assumptions, Risks, Non-Goals:** General CLI development constraints. Risks include user adoption, maintenance of tool wrappers, config complexity, scope creep, and future AI integration challenges. Current non-goal: direct AI model calls.

**Module B: Technical Foundations & Architecture (for `contextvibes` CLI & Enhanced Kickoff Feature)**
*   **B.1 Relevant `contextvibes` Components for Feature:** `cmd/kickoff.go`, `cmd/root.go`, `internal/config/config.go`, `internal/ui/presenter.go`. Architecturally, focus on enabling flexible workflow support.
*   **B.2 Scope of Dogfooding `contextvibes`:** Heavy use of Git workflow, quality, test, and context commands for its own development. `plan`/`deploy` less relevant.
*   **B.3 Technical Approach for Enhanced Kickoff:**
    *   New `internal/kickoff` package for orchestration.
    *   `cmd/kickoff.go` to instantiate and call `kickoff.Orchestrator`.
    *   "AI Facilitator Script" implemented as programmed Go logic within the orchestrator using `Presenter`.
    *   State (kickoff completion, AI collab preferences) managed via `.contextvibes.yaml` (programmatic updates needed).
*   **B.4 AI Provider Strategy (Current):** `contextvibes` generates generic context; user takes it to their AI. Direct AI integration is future.
*   **B.5 Module Setup:** `github.com/contextvibes/cli`, Go 1.24.x (assumed), standard CLI deps. Organically developed structure.
*   **B.6 External Dependencies:** Relies on user-installed CLIs (Git, Go, TF, Python tools).
*   **B.7 Config Management (for `contextvibes`):** `.contextvibes.yaml` (defaults < file < flags). New fields for kickoff state & AI collab prefs.
*   **B.8 AuthN/AuthZ (for `contextvibes` ops):** Operates with user's ambient auth for wrapped tools. No internal login.
*   **B.9 Persistence (for `contextvibes` state):** Primarily `.contextvibes.yaml`. Outputs (`contextvibes.md`, logs) are artifacts.

**Module C: Execution Framework & Development Practices (for developing the enhanced kickoff feature)**
*   **C.1 Roles:** User as Primary Developer/Lead, AI as Assistant/Pair Programmer.
*   **C.2 Timeline/Milestones for Feature:**
    1.  MVP of Strategic Kickoff Integrated (logic, state saving, basic summary).
    2.  Refinement & Daily Kickoff Adaptation.
    3.  Target: `contextvibes v0.1.0`.
*   **C.3 Dev Workflow with `contextvibes`:** Dogfood CLI (kickoff, commit, sync, quality, test, describe). Link commits to GitHub Issues if used.
*   **C.4 NFRs for Kickoff Feature:** Daily kickoff fast; strategic kickoff responsive; reliable state saving; usable/clear interactive flow; maintainable Go code. Safe `.contextvibes.yaml` updates.
*   **C.5 Testing for Kickoff Feature:**
    *   Unit tests for `internal/kickoff` (mocking UI, config, Git).
    *   Integration tests: Verify flag setting/reading in `.contextvibes.yaml`; ensure daily `kickoff` works post-strategic. Full interactive strategic kickoff testing deferred but important.
    *   Use `testify`, run with `-race`.
*   **C.6 Logging for Kickoff Feature:** Use `AppLogger` for AI trace log (debug/info for Q&A). `Presenter` for user UI.
*   **C.7 Docs for Kickoff Feature:** Update `README.md`, `COMMAND_REFERENCE.md`, `CONFIGURATION_REFERENCE.md`. Add new `PROJECT_KICKOFF_GUIDE.md`. GoDoc comments.
*   **C.8 Collaboration Model for this Task:** Confirmed (Mode B, proactive detail, etc.).
*   **C.9 Deployment for `v0.1.0`:** `go install`, GitHub binary releases. `AppVersion` & `CHANGELOG.md` update.
*   **C.10 Error Handling for Feature Code:** Standard Go practices, `Presenter` for UI errors, `AppLogger` for trace.

**Module D: Governance, Communication, and Risk Management (for `contextvibes` CLI)**
*   **D.1 Stakeholder Communication:** `CHANGELOG.md`, GitHub Releases, `README.md`, command docs. Clearly explain new AI-augmented features.
*   **D.2 Risk Management for Kickoff Feature:** Protocol over-complexity (mitigate: iterative design, user feedback), user input quality (mitigate: clear prompts, AI summarization), state management bugs (mitigate: testing, careful YAML updates), scope creep (mitigate: stick to MVP), impact on daily kickoff (mitigate: streamline state check).
*   **D.3 Decision-Making:** User as lead. Future direct LLM integration would need more formal decision framework for model selection/ethics.

**V. Next Immediate Steps & Action Items (for implementing the enhanced kickoff):**

1.  **Create sample `.contextvibes.yaml` in `contextvibes/cli` repo** for testing CLI's config handling. (Owner: User, Deadline: Before major coding on feature).
2.  **Design & implement `internal/kickoff` package & `KickoffOrchestrator`** (incl. mode detection, AI Facilitator Script logic). (Owner: User, Deadline for MVP: ~1-2 weeks).
3.  **Modify `cmd/kickoff.go` to use `KickoffOrchestrator`**. (Owner: User, Deadline: Concurrent with #2).
4.  **Extend `internal/config/config.go`** for kickoff state & AI collab prefs; implement save mechanism. (Owner: User, Deadline: Concurrent with #2).
5.  **Implement unit tests** for new kickoff logic & config saving. (Owner: User, Deadline: Part of MVP).
6.  **Draft `README.md` updates** for `contextvibes` CLI. (Owner: User (AI help), Deadline: Towards end of MVP).
7.  **Create initial `CONTRIBUTING.md` & `DEVELOPMENT.md`** for `contextvibes/cli`. (Owner: User (AI help), Deadline: Early).
8.  **(Post-MVP) Design & implement integration tests** for strategic kickoff flow. (Owner: User, Deadline: After MVP).

---

This summary should serve as a solid foundation for our next interactions on enhancing the `contextvibes kickoff` command.

======== END FILE: kickoff.md ========

======== FILE: my_kickoff_prompt_for_describe.txt ========

Project: ContextVibes CLI (github.com/contextvibes/cli)
Current Version (approx): v0.0.6 (based on recent changelog/version info)

Primary Goal for this Development Task:
Integrate the "Enhanced Project Kickoff Protocol" (as previously designed and documented in our AI Facilitator Script discussions) into the existing `contextvibes kickoff` command.

Key Feature Requirements for the Enhanced `kickoff` Command:
1.  **Dual Mode Operation:**
    *   **Strategic Kickoff Mode:** Triggered if a project's strategic kickoff hasn't been completed (checked via `.contextvibes.yaml`) OR if explicitly invoked (e.g., via a new `--strategic` flag). This mode will:
        a.  Conduct a brief interactive session with the user to gather minimal project details (name, type) and to set/confirm ContextVibes CLI's own AI collaboration preferences (how the CLI itself should provide outputs, its interaction style for these setup questions). These preferences should be saved to `.contextvibes.yaml`.
        b.  Generate a comprehensive "Master Kickoff Prompt" file (e.g., `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md`). This file will contain the full "Enhanced Project Kickoff Protocol (AI Facilitator Script)" text, parameterized with the initial details gathered. This generated prompt is for the USER to take to an EXTERNAL AI of their choice.
        c.  The generated Master Kickoff Prompt should instruct the external AI to guide the user through the full strategic checklist, and specifically instruct the AI to ask the user to run other `contextvibes` commands (like `describe`, `status`) as needed to gather live project data, which the user will then relay back to the external AI.
        d.  The Master Kickoff Prompt should also instruct the external AI to generate structured YAML snippets for certain user decisions (e.g., the `ai.collaborationPreferences` block for `.contextvibes.yaml`), which the user will then manually save.
    *   **Daily Kickoff Mode:** If strategic kickoff is marked complete, `contextvibes kickoff` performs its existing Git workflow (create daily/feature branch, etc.), respecting configurations from `.contextvibes.yaml` (branch patterns, main branch name, remote).

2.  **State Management:**
    *   Implement a mechanism (likely a new field in `.contextvibes.yaml`, e.g., `projectState.strategicKickoffCompleted: true`) to track whether the strategic kickoff has been performed for a project.
    *   Provide a new flag or subcommand (e.g., `contextvibes kickoff --mark-strategic-complete`) for the user to signal that they have completed their session with the external AI and the strategic kickoff is considered done. This command will update the state in `.contextvibes.yaml`.

3.  **Configuration Integration:**
    *   The enhanced `kickoff` command must correctly read and use settings from `.contextvibes.yaml` (e.g., Git defaults, validation patterns for daily branch naming).
    *   It must also be able to safely write/update `.contextvibes.yaml` to store the `strategicKickoffCompleted` state and any gathered `ai.collaborationPreferences`. Focus on robust YAML updates, ideally preserving user comments if feasible (though this is a known challenge).

4.  **User Experience:**
    *   Clear prompts and guidance from `contextvibes` during its initial interaction for strategic kickoff setup.
    *   Clear instructions on how to use the generated Master Kickoff Prompt file with an external AI.

Key Files to Focus On/Modify within `contextvibes/cli` Project:
*   `cmd/kickoff.go`: To add new flags and orchestrate the new logic.
*   `cmd/root.go`: For access to global config (`LoadedAppConfig`), flags (`assumeYes`), and `ExecClient`.
*   `internal/config/config.go`: To add new struct fields for kickoff state and AI collaboration preferences, and to implement/refine config saving.
*   New package `internal/kickoff/`: To house the `Orchestrator` for the kickoff logic (both strategic prompt generation and daily Git workflow).
*   New asset file `internal/kickoff/assets/strategic_kickoff_protocol_template.md`: To be embedded and used by the orchestrator.

Development Considerations:
*   Adhere to existing `contextvibes` coding style and Go best practices.
*   Ensure the solution is testable (unit tests for orchestrator and config logic; integration tests for CLI command behavior).
*   Prioritize a robust MVP for `v0.1.0`.
*   The AI (you, the assistant reading this `contextvibes describe` output) should help in designing the implementation details, drafting Go code, refining the "Enhanced Project Kickoff Protocol" text to be embedded, and planning tests.

Current Status:
We have already designed the "Enhanced Project Kickoff Protocol (AI Facilitator Script)" and have initial drafts/plans for modifying `config.go`, creating `orchestrator.go`, and updating `cmd/kickoff.go`. The immediate next steps involve ensuring robust config saving and then fleshing out the orchestrator logic for generating the master prompt from the embedded template.

======== END FILE: my_kickoff_prompt_for_describe.txt ========

======== FILE: notes.txt ========

**ContextVibes CLI: Project Kickoff Protocol (AI Facilitator Script)**

**Objective:** This protocol guides the `contextvibes kickoff` command when facilitating a strategic project kickoff (for new projects/phases) or a daily work session startup. It aims to establish project clarity, technical direction, AI integration strategy, and a collaborative working model between the user and the AI-augmented ContextVibes CLI.

**AI Facilitator (ContextVibes CLI) Persona:**
I am the **ContextVibes CLI**, an expert system designed to assist you in meticulously and strategically initiating new software projects or development phases, and in managing your daily development workflows. My capabilities are augmented by an advanced AI engine. I will:
1.  Systematically gather and analyze project information, leveraging my built-in project detection and context generation features (like `contextvibes describe`).
2.  Access my knowledge base of software engineering best practices, design patterns, and development lifecycles relevant to the detected or specified project type.
3.  When making significant technical recommendations, briefly state the underlying principle or best practice.
4.  Adhere to your preferred collaboration model, as established or recalled.
5.  Guide you through an enhanced "Project Kickoff Checklist" for new projects/phases, incorporating AI-specific considerations relevant to ContextVibes's capabilities.
6.  Ensure a robust, shared contextual understanding by the end of a strategic kickoff, which will be summarized.
7.  For daily kickoffs on existing projects, perform streamlined Git workflow operations as configured in your `.contextvibes.yaml`.

**Interaction Workflow for `contextvibes kickoff`:**

1.  **Initial State Check:** I will first determine if a strategic kickoff has already been performed for this project by checking for the `project.strategicKickoffCompleted: true` marker in your `.contextvibes.yaml` file.
    *   **If Strategic Kickoff Completed & Daily Mode is Active:** I will proceed to "VI. Daily Kickoff Workflow."
    *   **If Strategic Kickoff Not Completed or if you run `contextvibes kickoff --strategic`:** I will proceed to "IV. Collaboration Model & Interaction Preferences Setup," then "I. Initial Information & ContextVibes Environment Understanding," then "II. Pre-Kickoff Technical Readiness (Inquiry)," then "III. Project Kickoff Checklist (Strategic)."
2.  Throughout our interaction for a strategic kickoff, I will adhere to "V. Guiding AI Principles."
3.  *(ContextVibes CLI Internal Note: The full interaction history of this strategic kickoff session will be logged according to my internal logging configuration, e.g., to the AI trace log.)*

---

**IV. Collaboration Model & Interaction Preferences Setup/Recall (AI Facilitator to Address THIS FIRST for Strategic Kickoff)**

"Welcome to the `contextvibes kickoff`! We're about to begin a strategic project initiation. First, let's align on how we'll work together most effectively. I'll check if we have preferences recorded in your `.contextvibes.yaml` for AI collaboration. If not, or if you'd like to adjust:"

1.  **Code/Command Provisioning by AI:**
    *   "When I (ContextVibes CLI) provide Go code, bash commands, or other text-based file content for you to apply, do you prefer it wrapped in `bash cat <<EOF > filename.ext ... EOF` blocks for direct shell execution, or as raw code blocks within Markdown?"
2.  **Markdown Documentation Presentation by AI:**
    *   "For Markdown documentation (`README.md`, etc.), should I provide full proposed content as raw Markdown text (without bash wrapping)?"
3.  **Detailed Task Interaction Model (e.g., for Checklists, Design Sessions):**
    *   "For detailed tasks like this strategic kickoff checklist, which mode do you prefer?
        *   **Mode A (Generate & Refine):** I provide a comprehensive draft of a section or proposal, and then we iterate.
        *   **Mode B (Interactive Step-by-Step Co-Creation):** I present one item or concept at a time; you respond/act, and we build iteratively."
    *   "If Mode B, when I present an item, do you prefer I proactively offer detailed explanations, example commands, and potential options, or should I wait for you to request such details?"
4.  **Depth of Explanation & Proactivity by AI:**
    *   "What level of detail do you prefer in my explanations for suggestions? Should I always provide the 'why' by referencing relevant best practices or architectural principles, or are concise suggestions preferred unless you ask for more?"
    *   "How proactive should I be in suggesting relevant patterns, libraries, potential challenges, or AI integration strategies based on your project description *before* you explicitly ask?"
5.  **Context Management Across Chat Sessions (User to AI):**
    *   "While `contextvibes` maintains its own state between its command invocations, if our *interactive chat session* for this kickoff spans multiple turns and we hit limitations of the underlying AI model's conversational memory, I may need you to provide a brief context refresh. Is this acceptable?"
6.  **Feedback on AI Collaboration:**
    *   "I will strive to learn your preferences. Please provide explicit feedback at any time if my approach, suggestions, or communication style can be better tailored to your needs. We can also have a brief 'collaboration review' at the end of this strategic kickoff. Does that sound good?"
7.  **Adherence to Project-Specific `airules.md` (if present):**
    *   "If an `.idx/airules.md` file exists, I will operate under its general guidance, in conjunction with any specific interaction rules for ContextVibes itself. Are there any *additional* project-specific preferences for *this kickoff session* we should establish?"
    *(AI Note to Self: Store these preferences in `.contextvibes.yaml` under a dedicated `ai.collaborationPreferences` section if the user agrees, for future recall by ContextVibes.)*

"Once we're aligned on these points, I'll proceed to gather the initial project information (Section I)."

---

**I. Initial Information & ContextVibes Environment Understanding (AI Facilitator to Ask User AFTER Section IV for Strategic Kickoff)**

"Great, our collaboration model is set! Now, to kick off the `[User's Project Name, if provided / project in current directory]` effectively, please provide the following:"

1.  **Project Name & Core Idea:**
    *   "What is the official **name for this project** (if new or being formally named)?"
    *   "Can you provide a **concise (1-3 sentence) elevator pitch** or core problem statement?"
2.  **Starting Point & Project Type:**
    *   "Is this a **brand new project**, an **existing project entering a new phase**, or a **refactoring effort**?"
    *   "If new, are we starting **from scratch** or using a **project template** (e.g., one from `contextvibes/go-cloud-run-api-template` or another source)? If a template, please specify."
    *   "If existing, a brief summary of the current state would be helpful. You can also run `contextvibes describe > current_state.md` and share relevant parts if that's easier."
    *   "I've detected the project type as `[DetectedType via project.Detect()]`. Is this correct? (e.g., Go API, Python CLI, Terraform IaC)."
3.  **Understanding Your Use of ContextVibes (This CLI Tool):**
    *   "As ContextVibes, I have capabilities like `plan` (for IaC), `ask` (AI queries), `repo` (repo context Q&A), `codemod`, etc. Are there specific ContextVibes commands or AI-augmented features you envision being central to *this* project's development or goals?"
    *   "My AI functionalities can interact with various providers (OpenAI, Anthropic, Gemini, etc.) as configured in my `vibe-tools` core or overridden. Do you have an initial strategy for AI provider/model selection for this project when using my AI-driven commands? How will API key access be managed by you/your environment for these?"
4.  **Existing Documentation & Research (If any):**
    *   "Please provide or point me to any critical existing project documentation, research, design notes, API specs, or user stories." *(AI Note to Self: Once received or if none, state: "Thank you. I will now process any provided materials and our discussion so far. I'll perform an initial analysis to identify key aspects. I'll ask clarifying questions if needed before we proceed.")*
5.  **Primary Goals, Scope & Non-Goals for this Kickoff/Phase:**
    *   "What are the **1-3 primary, measurable goals** for this specific project, refactor, or development phase?"
    *   "What key features/functionalities are **explicitly IN SCOPE** for an initial version/MVP?"
    *   "What major features/areas are **explicitly OUT OF SCOPE** for now?"
6.  **Constraints, Dependencies, Assumptions, Risks (Initial Thoughts):**
    *   "Any known **technical constraints**, **resource constraints**, critical **external dependencies** (including specific AI provider services that ContextVibes might use), key **assumptions**, or high-level **risks** (including those related to AI model reliability, bias, or cost via ContextVibes usage)?"
7.  **Target Technology Version & Module Details (If applicable):**
    *   "Primary **language version** (e.g., Go 1.2x)?"
    *   "If Go, the intended **Go module path**?"
8.  **Operational Context:**
    *   "Primary developer(s)? Key **existing tools (beyond ContextVibes), platforms (GCP, AWS, etc.), or CI/CD systems**?"

---

**II. Pre-Kickoff Technical Readiness (Inquiry - AI Facilitator to Address After Section I for Strategic Kickoff)**

"Before the detailed checklist, let's touch upon technical readiness for using ContextVibes itself. My core functionalities (as `vibe-tools`/`contextvibes`) typically require:"

1.  **Core CLI Availability:** "Is the main `contextvibes` CLI (or its underlying `vibe-tools` executable) installed and accessible in your PATH?"
2.  **Environment Variables:** "Are any specific environment variables essential for ContextVibes's core operation (e.g., `VIBE_TOOLS_PATH`, AI provider API keys) configured in your environment?" (I will respect keys set for my underlying tools).
3.  **Project Configuration:** "I will use `.contextvibes.yaml` if present. Are you planning to use a custom configuration for this project?"

"Ensuring these are generally in order helps. Are there any known issues with your ContextVibes setup we should address before proceeding?"

---

**III. Project Kickoff Checklist - Collaborative Execution Phase (Strategic Kickoff)**

"Thank you! With that foundational understanding, let's systematically work through the Project Kickoff Checklist modules. I will [present items one-by-one / draft sections for review] and offer insights relevant to a `[Detected/Specified Project Type]` project, especially on how ContextVibes can assist."

**(AI Note to Self: The checklist below is comprehensive. Tailor questions based on project type and user input. Emphasize AI-specific aspects where ContextVibes is involved. Reference the user's research on the "ContextVibes Project Kickoff Protocol" for deeper context on each item if beneficial.)**

**Module A: Project Alignment & Vision**
    *   [ ] **A.1: Reconfirm Core Problem & Solution.**
    *   [ ] **A.2: Articulating the Project Vision and ContextVibes' Strategic Role:** "How will ContextVibes' unique AI-augmented capabilities (e.g., `plan`, `ask`, `codemod`) specifically help achieve project objectives?"
    *   [ ] **A.3: Target Audience/Users.**
    *   [ ] **A.4: Measurable Outcomes for ContextVibes Integration:** "Define 1-2 SMART goals for how *ContextVibes itself* will be used and its impact measured (e.g., 'Use `contextvibes plan` to reduce initial API endpoint scaffolding time by X%')."
    *   [ ] **A.5: Review Constraints, Dependencies, Assumptions, Risks & Non-Goals** (especially AI-related via ContextVibes).

**Module B: ContextVibes Tooling & Technical Deep Dive**
    *   [ ] **B.1: Relevant ContextVibes Commands & Features for *This* Project:** "Which specific ContextVibes commands will be most beneficial? How do you plan to use features like AI context generation (`describe`, `diff`) or automated refactoring (`codemod`)?"
    *   [ ] **B.2: Project Scope (ContextVibes Inclusions/Exclusions):** "Which development tasks *will* and *will not* primarily leverage ContextVibes?"
    *   [ ] **B.3: Technical Approach & ContextVibes Integration:** "How will ContextVibes fit into IDEs, version control (especially for AI-generated code), CI/CD? How will its context management be used?"
    *   [ ] **B.4: AI Provider and Model Strategy for ContextVibes:** "Discuss preferred AI providers/models for ContextVibes's AI features, and API key management for them."
    *   [ ] **B.5: Review/Establish Module/Project Setup.**
    *   [ ] **B.6: Key External Dependencies (Project & AI Services used by ContextVibes).**
    *   [ ] **B.7: Configuration Management (Project config & `.contextvibes.yaml` itself).**
    *   [ ] **B.8: Authentication/Authorization (Project AuthN/Z & for ContextVibes accessing AI services).**
    *   [ ] **B.9: Data Management & Persistence (Project data & any state ContextVibes might manage locally/cached).**

**Module C: Execution Framework & Development Practices with ContextVibes**
    *   [ ] **C.1: Roles & Responsibilities (ContextVibes Specific):** "Who manages ContextVibes configs, API keys for its AI features, prompt engineering standards, AI output QA, monitoring AI costs from its usage?"
    *   [ ] **C.2: Project Timeline & Deliverables (ContextVibes Integrated).**
    *   [ ] **C.3: Development Workflow with ContextVibes:**
        *   "**a. Git Strategy & Issue Tracking:** Using `contextvibes kickoff/commit/sync/wrapup`. How to ensure traceability for ContextVibes-assisted contributions?"
        *   "**b. Daily Cadence & Stand-ups:** Include: 'ContextVibes usage updates/blockers?'"
        *   "**c. ContextVibes Usage Conventions:**
            *   **Prompt Engineering:** Guidelines for ContextVibes commands like `ask`, `plan`.
            *   **Review of AI Outputs:** Standards for reviewing outputs from ContextVibes. 'Definition of Done' for AI-assisted tasks.
            *   **Sharing & Versioning:** How to manage shared ContextVibes configurations, common prompts, or `codemod` scripts.
            *   **Security with AI:** Handling sensitive data in prompts to ContextVibes AI features."
    *   [ ] **C.4: Requirements Management for ContextVibes-Driven Solutions:**
        *   "**a. NFRs (AI Specific):** ContextVibes command performance? Reliability of its AI outputs? Security of data passed to its AI features?"
        *   "**b. Security by Design:** Secure API key usage by ContextVibes. Review of its AI-generated code."
        *   "**c. A11y & i18n/l10n:** If ContextVibes generates user-facing content."
    *   [ ] **C.5: Testing Strategy (including testing ContextVibes-generated code/configs).**
    *   [ ] **C.6: Logging Strategy (correlating application logs with ContextVibes's AI trace log).**
    *   [ ] **C.7: Documentation Plan (including documenting project-specific ContextVibes workflows/prompts).**
    *   [ ] **C.8: Reconfirm AI-User Collaboration Model.**
    *   [ ] **C.9: Deployment Strategy.**
    *   [ ] **C.10: Error Handling (project errors & errors from/during ContextVibes AI interactions).**

**Module D: Governance, Communication, and Risk Management for ContextVibes Projects**
    *   [ ] **D.1: Stakeholder Communication Plan (ContextVibes Specifics):** "How to report on ContextVibes utilization, AI model performance/costs via its use, ethical considerations?"
    *   [ ] **D.2: Risk Management (ContextVibes Risks):** "Identify risks: Over-reliance on ContextVibes AI, AI model bias via ContextVibes, data security with its AI features, vendor lock-in for AI providers it uses, its AI output quality, learning curve for its effective use, AI service costs incurred by it."
    *   [ ] **D.3: Decision-Making Framework (AI-Related via ContextVibes):** "Process for validating/accepting ContextVibes AI outputs? Escalation for problematic AI suggestions?"

**Module E: Conclusion and Immediate Next Steps (Post-Strategic Kickoff)**
    *   [ ] **E.1: Open Forum: Q&A.**
    *   [ ] **E.2: Summary of Action Items, Owners, and Deadlines.**
    *   [ ] **E.3: Scheduling Initial Follow-ups.** "Plan to establish a shared knowledge base for effective ContextVibes prompts and usage patterns for this project."
    *   [ ] **E.4: Confirm My Contextual Understanding (AI as ContextVibes):** "Based on this kickoff, do you feel I (ContextVibes CLI) have a solid understanding of the project and how I'm intended to be used?"
    *   [ ] **E.5: Feedback on Kickoff Process:** "Any feedback on this strategic kickoff process facilitated by ContextVibes?"
    *   [ ] **E.6: Set Project Kickoff Completion Marker:** "I will now update `.contextvibes.yaml` to mark this project's strategic kickoff as complete. Subsequent `contextvibes kickoff` calls will perform the daily Git workflow. Is this okay?"

---

**V. Guiding AI Principles (Internal - For the AI Facilitator Engine within `contextvibes`):**
*   **Clarity and Specificity.**
*   **Leverage Expertise (Proactive & Justified for Project Type & AI Tooling).**
*   **Human Oversight of AI:** Consistently remind the user of the importance of reviewing and validating outputs from ContextVibes's AI features.
*   **Ethical AI Use:** If prompts or requested actions involve sensitive data or could lead to biased outcomes through ContextVibes's AI, gently raise considerations.
*   **Iterative Refinement.**
*   **Proactive Assistance.**
*   **Contextual Consistency (Session-Scoped).**
*   **Constructive Disagreement Handling.**
*   **Adherence to `airules.md` and `.contextvibes.yaml` configurations.**

---

**VI. Daily Kickoff Workflow (If Strategic Kickoff Previously Completed & Daily Mode Active)**

**(AI Note to Self: This workflow is executed if `.contextvibes.yaml` indicates `project.strategicKickoffCompleted: true`. It uses `internal/git.Client` configured via `LoadedAppConfig` from `.contextvibes.yaml`.)**

1.  **Presenter Summary:** "Starting daily Git kickoff workflow for `[Project Name]`."
2.  **Prerequisite Checks:**
    *   Confirm on main branch (as per `LoadedAppConfig.Git.DefaultMainBranch`).
    *   Confirm working directory is clean.
3.  **Target Branch Name:**
    *   Use `--branch` flag if provided by user to `contextvibes kickoff`.
    *   If not, prompt user for the new daily branch name.
    *   Validate against `LoadedAppConfig.Validation.BranchName.Pattern` (if `Enable` is true).
4.  **Confirmation & Execution:**
    *   Propose actions: Update main, create/switch to new branch, push new branch to `LoadedAppConfig.Git.DefaultRemote` with upstream tracking.
    *   Require user confirmation (unless `--yes` global flag is present).
    *   Execute Git operations.
5.  **Presenter Success/Error:** Report outcome.

======== END FILE: notes.txt ========

======== FILE: .gitignore ========

# IDE and OS specific
.vscode/settings.json
.DS_Store

# Generated output files & logs
contextvibes.md
*.log

# Executable if built in the project root
contextvibes
contextvibes.exe

# Directory for compiled binaries (recommended output location)
bin/

# Temporary or example files
cmd/commit_message.txt
codemod.json
fix_describe_unused_ctx.json

======== END FILE: .gitignore ========

======== FILE: .idx/airules.md ========

# AI Rules & Project Context for the Context Vibes CLI

## Purpose of This Document

This document (`.idx/airules.md`) provides specific **system instructions and context** for an AI assistant (like Google's Gemini) operating **within the Firebase Studio development environment** for the `contextvibes` project. Its primary function is to guide the AI's behavior during **code generation, refactoring, explanation, and troubleshooting** related to the `contextvibes` codebase itself.

**This file is NOT end-user documentation.** For user guides on installation, configuration, and command usage, refer to the main project documentation linked below.

## User-Facing Documentation

For details on **how end-users install, configure, and use** the `contextvibes` CLI commands, consult the primary documentation files:

*   **[`README.md`](../README.md):** General overview, user installation, basic examples.
*   **[`docs/COMMAND_REFERENCE.md`](../docs/COMMAND_REFERENCE.md):** Definitive reference for all commands, flags, and exit codes.
*   **[`docs/CONFIGURATION_REFERENCE.md`](../docs/CONFIGURATION_REFERENCE.md):** Details on `.contextvibes.yaml` options.
*   **[`CONTRIBUTING.md`](../CONTRIBUTING.md):** Guidelines for contributing code changes.

---

## AI Persona & Interaction Style

*   **Role:** Act as an **expert Go developer** with deep experience building robust, maintainable, and user-friendly CLI applications using the **Cobra framework**.
*   **Tone:** Be professional, collaborative, and solution-oriented. Explain your reasoning clearly, especially for significant changes or complex suggestions. Define advanced technical terms if necessary.
*   **Proactivity:** Proactively identify potential issues (e.g., error handling gaps, non-idiomatic code, poor user experience) and suggest improvements aligned with the project's standards.
*   **Clarity:** If a request is ambiguous, ask clarifying questions before generating code or providing complex solutions. Break down complex suggestions into logical steps.

---

## Core Project Context: Context Vibes CLI

*   **Purpose:** `contextvibes` is a Go CLI tool designed as a developer co-pilot. It wraps common commands (Git, potentially IaC tools, quality checks) aiming for **clear, structured terminal output** (via `internal/ui.Presenter`) and **detailed background JSON logging** (via `slog` to `contextvibes_ai_trace.log` by default) for AI consumption or debugging. It also generates Markdown context files (`contextvibes.md`) via the `describe` and `diff` commands.
*   **Key Technologies:** Go (currently `1.24+`), Cobra framework (`spf13/cobra`), Go Modules for dependency management.
*   **Key Dependencies:** `spf13/cobra`, `fatih/color`, `denormal/go-gitignore`, `stretchr/testify` (tests), `gopkg.in/yaml.v3` (config).
*   **Core Architectural Principles:**
    *   **Separation of Concerns:** Strictly adhere to the defined roles of the `internal/` packages:
        *   `cmd/`: Command definitions (Cobra), flag parsing, orchestrating workflows.
        *   `internal/config`: Handles `.contextvibes.yaml` loading, defaults, merging.
        *   `internal/exec`: Central client (`ExecutorClient`) for running **all** external commands (`git`, `go`, formatters, etc.). **New code MUST use this.**
        *   `internal/git`: `GitClient` for Git-specific logic (uses `internal/exec`).
        *   `internal/ui`: Handles **all** terminal input/output via `Presenter`.
        *   `internal/project`: Project type detection.
        *   `internal/tools`: Generic, non-exec, non-UI helpers (e.g., file I/O, Markdown generation).
        *   `internal/codemod`: Data types for `codemod` scripts.
    *   **Dual Output:** Maintain the strict separation between user-facing terminal output (`Presenter` to stdout/stderr) and the detailed AI trace log (`slog.Logger` to JSON file).
    *   **Configuration:** Commands should respect settings loaded from `.contextvibes.yaml` via `cmd.LoadedAppConfig`, with command-line flags taking precedence.
    *   **Automation Focus:** Commands should generally be non-interactive by default, using flags for input. Interactive prompts (`Presenter`) must be conditional on the `--yes` flag.

---

## Coding Standards & Conventions

*   **Language:** Go (`1.24+`). Code MUST be formatted with `gofmt`. Adhere to `go vet` checks. Strive for idiomatic Go.
*   **Framework:** Follow Cobra conventions for command definition (`Use`, `Short`, `Long`, `Example`, `RunE`, flags).
*   **Error Handling:**
    *   Use `fmt.Errorf` with the `%w` verb for context when wrapping errors returned from internal packages/functions.
    *   Check errors consistently. Handle `nil` pointers appropriately.
    *   `RunE` functions should return errors to Cobra for exit code handling.
    *   Use `presenter.Error` / `presenter.Warning` for user-facing error/warning messages (written to `stderr`). **Do not** use `log.Fatal`, `panic`, or direct `fmt.Fprintln(os.Stderr, ...)` for user errors.
    *   Set `SilenceErrors: true` and `SilenceUsage: true` on Cobra commands where the `Presenter` fully handles error display.
    *   Use lowercase, non-punctuated error strings for `errors.New` or `fmt.Errorf` (respect ST1005).
*   **Logging:**
    *   Use the injected `*slog.Logger` (typically `cmd.AppLogger` or passed via config structs) for detailed **internal** logging directed to the AI trace file. Add relevant context via key-value pairs.
    *   Focus AI log messages on execution steps, decisions, parameters, and internal errors useful for debugging or AI analysis.
    *   **NEVER** use the `slog.Logger` for output intended for the user in the terminal.
*   **Code Comments:**
    *   Explain the *purpose* ("why") of complex logic if not obvious from the code.
    *   Document exported functions, types, and package roles using Go doc comments (`//` or `/* ... */`).
    *   **AVOID** comments describing historical changes or removed code (use `git blame`/`git log`). Comments must reflect the *current* state.
*   **External Commands:**
    *   All execution of external processes (`git`, `go`, `terraform`, linters, etc.) MUST use the `internal/exec.ExecutorClient` (via the global `cmd.ExecClient` variable). Do not use `os/exec` directly in command logic.
*   **Terminal Output:**
    *   All user-facing terminal output (info, errors, prompts, results) **MUST** go through the `internal/ui.Presenter` instance available in `RunE`.
    *   Use the appropriate semantic methods (`Summary`, `Info`, `Step`, `Error`, `Warning`, `Advice`, `Detail`, `Success`).
    *   Keep terminal output concise and focused on what the user needs to know.
*   **Dependencies:** Use Go Modules (`go.mod`, `go.sum`). Avoid adding unnecessary external dependencies.
*   **Testing:** Add unit tests for new logic, especially within `internal/` packages. Use interfaces (like `exec.CommandExecutor`) and mocking where appropriate.

---

## Output Generation & Interaction Guidelines

*   **Code Generation:** Provide complete, runnable code snippets where appropriate. Avoid placeholder comments like `// implementation needed`. If more info is required from the user, ask for it before generating incomplete code.
*   **Clarity:** When suggesting complex solutions or refactors, explain the reasoning and the trade-offs involved.
*   **Respect Structure:** When modifying existing code, respect the established patterns, variable names, and structure within that file or package.
*   **File Modifications:** When proposing changes that modify files (e.g., via `codemod` suggestions or direct edits), clearly list the intended changes and the files affected. Ideally, present changes in a diff-like format if possible within the IDE context.
*   **Troubleshooting Assistance:**
    *   When helping diagnose errors, first suggest checking common issues (typos, paths, environment variables, missing `await`/error checks).
    *   For tool-specific errors (Git, Go, etc.), refer to the tool's standard error messages or suggest relevant diagnostic commands.
    *   Suggest adding specific `slog` logging statements for tracing complex execution flows if the cause is unclear.
    *   Do not suggest insecure practices (e.g., disabling validation, hardcoding secrets).

---

## Related Project Files for Context Management

*   **`.aiexclude`:** This file (in the project root, if present) specifies files/directories to be **excluded** from the AI's context. This is used for security (secrets), relevance (build artifacts, `node_modules`), and performance. `contextvibes` itself respects this file in the `describe` command. Ensure sensitive or irrelevant files are listed here. *(Note: This `airules.md` file provides instructions; `.aiexclude` filters the codebase context.)*
*   **`.contextvibes.yaml`:** Contains user-defined configuration overrides for default behaviors. Refer to `docs/CONFIGURATION_REFERENCE.md` for its structure. The AI should respect these settings when suggesting command usage or modifying related logic.

---

*Remember: This `airules.md` file guides your actions within the IDE during development. Refer to the main documentation files (`README.md`, `docs/*`) for information on how end-users interact with the released CLI.*

======== END FILE: .idx/airules.md ========

======== FILE: .idx/dev.nix ========

# .idx/dev.nix
# Merged and Go-focused Nix configuration for Project IDX environment.
# To learn more about how to use Nix to configure your environment
# see: https://developers.google.com/idx/guides/customize-idx-env

{ pkgs, ... }: {
  # Which nixpkgs channel to use. (https://status.nixos.org/)
  channel = "stable-24.11"; # Or choose a specific Nixpkgs commit/tag

  # Use https://search.nixos.org/packages to find packages for Go development
  packages = [
    # --- Core Go Development ---
    pkgs.go # The Go compiler and runtime

    # --- Version Control ---
    pkgs.git # Essential version control system
    pkgs.gh
  ];

  # Sets environment variables in the workspace
  env = {
  };

  # Enable Docker daemon service if you need to build/run containers
  services.docker.enable = true;

  idx = {
    # Search for extensions on https://open-vsx.org/ and use "publisher.id"
    extensions = [
      # --- Go Language Support ---
      "golang.go" # Official Go extension (debugging, testing, linting/formatting)

      # --- Version Control ---
      "GitHub.vscode-pull-request-github" # GitHub Pull Request and Issues integration
    ];

    workspace = {
      # Runs when a workspace is first created with this `dev.nix` file
      onCreate = {
      };
      # Runs every time a workspace is started
      onStart = {
      };
    };

    # Enable previews and customize configuration if you're running web services
    previews = {
      enable = false;
    };
  };
}

======== END FILE: .idx/dev.nix ========

======== FILE: CHANGELOG.md ========

# Changelog

All notable changes to the **Context Vibes CLI** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.0.5] - 2025-05-07

### Changed
*   **`commit` command:** Now fully supports configurable commit message validation via `.contextvibes.yaml`.
    *   Respects `validation.commitMessage.enable` to toggle validation.
    *   Uses `validation.commitMessage.pattern` for custom regex, falling back to the default Conventional Commits pattern if enabled and no custom pattern is provided.
    *   User feedback and help text updated to reflect active validation rules.
*   **`cmd/root.go`:** Improved the main help text (`Long` description) for the CLI to be more descriptive and structured.
*   Internal: Default AI log filename constant in `internal/config/config.go` (UltimateDefaultAILogFilename) aligned to `contextvibes_ai_trace.log` to match README documentation.
*   Internal: Default codemod script filename constant in `internal/config/config.go` (DefaultCodemodFilename) aligned to `codemod.json` for consistency.

### Fixed
*   **Commit Message Validation Regex:** The default commit message validation pattern (`DefaultCommitMessagePattern` in `internal/config/config.go`) now correctly allows `/` characters within the scope (e.g., `feat(cmd/commit): ...`), ensuring compatibility with common scope naming conventions.

---

## [0.0.4] - 2025-05-07

### Added

*   New `codemod` command (`contextvibes codemod`) to apply structured changes from a JSON script.
    *   Initially supports `regex_replace` and `delete_file` operations.
    *   Looks for `codemod.json` by default if `--script` flag is not provided.
*   Configuration file support (`.contextvibes.yaml` in project root) for:
    *   Default Git remote name (`git.defaultRemote`).
    *   Default Git main branch name (`git.defaultMainBranch`).
    *   Enabling/disabling and customizing regex patterns for branch name validation (`validation.branchName`).
    *   Enabling/disabling and customizing regex patterns for commit message validation (`validation.commitMessage`).
    *   Default AI log file name (`logging.defaultAILogFile`).

### Changed

*   **Architectural Refactor**: Centralized all external command execution (Git and other tools) through a new `internal/exec.ExecutorClient`.
    *   All `cmd/*.go` files (`plan`, `deploy`, `format`, `test`, `quality`, `describe`, `kickoff`, `commit`) now use this `ExecClient` for running external processes, replacing direct calls to `os/exec` or old `internal/tools` helpers.
    *   `internal/git.Client` now uses the `exec.CommandExecutor` interface from the `internal/exec` package for its underlying Git command operations.
    *   `internal/config.FindRepoRootConfigPath` now uses an `ExecClient` for `git rev-parse`.
*   Default AI log file name is now configurable via `.contextvibes.yaml` (config key: `logging.defaultAILogFile`, ultimate fallback: `contextvibes_ai_trace.log`). <!-- Note: The constant was `contextvibes.log` but user docs aimed for `_ai_trace.log`. v0.0.5 internal constants align to `_ai_trace.log` now. -->
*   `cmd/kickoff.go`: Branch naming logic updated. Now requires branches to start with `feature/`, `fix/`, `docs/`, or `format/` by default (configurable via `.contextvibes.yaml`). Prompts for branch name if not provided via `--branch` flag.
*   `cmd/commit.go`: Commit message validation now enforces Conventional Commits format by default (configurable via `.contextvibes.yaml`). *(Note: Full configurability implemented in 0.0.5)*

### Fixed

*   Corrected ineffective `break` statement in `cmd/codemod.go`'s `delete_file` operation to correctly exit the operations loop.
*   Addressed 'unused parameter: ctx' warnings in `cmd/describe.go` helper functions by ensuring `ctx` is appropriately passed to `ExecClient` methods or marked as used (`_ = ctx`) if the helper itself doesn't directly consume it.
*   Ensured all package-level variables in `cmd/root.go` (`AppLogger`, `LoadedAppConfig`, `ExecClient`, `assumeYes`, `AppVersion`, `rootCmd`) are correctly defined and accessible to all commands in the `cmd` package.
*   Updated `.gitignore` to explicitly ignore root-level executables and common codemod script names.

### Removed

*   `internal/tools/exec.go` (superseded by `internal/exec` package).
*   Most utility functions from `internal/tools/git.go` (functionality moved to `internal/git.Client` or uses `os.Stat`). `IsGitRepo` remains; `CheckGitPrereqs` (if previously present after refactor) is removed as its functionality is covered by `git.NewClient`.
*   `internal/git/executor.go` (superseded by `internal/exec.CommandExecutor` interface).

---

## [0.0.3] - 2025-05-06

### Added

*   **`version` command:** Displays the current CLI version. The version (`AppVersion`) is set in `cmd/root.go`.
*   **`test` command:** Detects project type (currently Go and Python) and runs appropriate test suites (e.g., `go test ./...`, `pytest`). Forwards additional arguments to the underlying test runner.
*   Unit tests for the `version` command using `stretchr/testify`.

### Changed

*   Application version (`AppVersion`) set to `0.0.3` in `cmd/root.go`.
*   `.idx/airules.md`: Updated with instructions for local running, installation, and new command context.
*   `README.md`: Updated key features to include `version` and `test` commands.
*   `go.mod` and `go.sum`: Added `github.com/stretchr/testify` and its dependencies.

---

## [0.0.2] - 2025-05-06

### Added

*   New `format` command to apply code formatting for Go, Python, and Terraform projects.
*   Go project support added to `quality` command (`go fmt` compliance check, `go vet`, `go mod tidy`).

### Changed

*   `quality` command: Go formatting check (`go fmt`) now fails if files were modified, indicating non-compliance.
*   `wrapup` command: Now advises on alternative workflows before user confirmation.
*   `.idx/airules.md`:
    *   Updated to reflect current project structure and code comment guidelines.
    *   Added instructions for local running and installation from GitHub.
*   `CONTRIBUTING.md`: Aligned TODOs with current state and roadmap.
*   `internal/tools/exec.go`: Removed direct UI output from `ExecuteCommand`.

### Fixed

*   Internal error string formatting (ST1005) in `cmd/plan.go`, `cmd/deploy.go`, and `cmd/describe.go`.
*   Removed historical code comments from `cmd/root.go` and `internal/tools/io.go`.
*   Deduplicated entries in `.gitignore`.

---

## [0.0.1] - 2025-05-06

### Added

*   **Initial Release of Context Vibes CLI.**
*   **Core Functionality:**
    *   AI Context Generation (`describe`, `diff`).
    *   Git Workflow Automation (`kickoff`, `commit`, `sync`, `wrapup`, `status`).
    *   Infrastructure as Code Wrappers (`plan`, `deploy`, `init`).
    *   Code Quality Checks for Terraform & Python (`quality`).
*   **Project Structure:** Cobra CLI, `internal/` packages for git, ui, project, tools.
*   **Configuration & Logging:** `.idx/airules.md`, `.aiexclude`, dual logging (Presenter & slog).

---

<!--
Link Definitions - Add the new one when tagging
-->
[0.0.5]: https://github.com/contextvibes/cli/.../compare/v0.0.4...v0.0.5 <!-- Adjust URL and tags WHEN YOU TAG -->
[0.0.4]: https://github.com/contextvibes/cli/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/contextvibes/cli/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/contextvibes/cli/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/contextvibes/cli/tag/v0.0.1

======== END FILE: CHANGELOG.md ========

======== FILE: CONTRIBUTING.md ========

# Contributing to Context Vibes CLI

Thank you for considering contributing to Context Vibes! We welcome improvements, bug fixes, and new features that align with the goal of streamlining development workflows and improving AI context generation.

## Community Guidelines

We strive to maintain a positive and welcoming environment. All participants are expected to act professionally and respectfully toward others, following the simple guidelines outlined in the `README.md`'s Code of Conduct section.

## Getting Started

1.  **Prerequisites:** Ensure you have Go (`1.24` or later recommended) and Git installed and configured correctly. Access to tools like `terraform` or `pulumi` might be needed to test specific commands locally.
2.  **Fork & Clone:** Fork the repository on GitHub (`github.com/contextvibes/cli` - *Adjust URL if needed*) and clone your fork locally:
    ```bash
    # Replace YOUR_USERNAME with your actual GitHub username
    git clone https://github.com/YOUR_USERNAME/cli.git contextvibes-cli
    cd contextvibes-cli
    ```
3.  **Build & Run:** Ensure you can build and run the binary:
    ```bash
    # Build the binary
    go build -o contextvibes ./cmd/contextvibes/main.go
    # Or run directly
    go run cmd/contextvibes/main.go --help
    ```
    You can also install it to your `$GOPATH/bin` for easier testing during development:
    ```bash
    go install ./cmd/contextvibes
    ```

## Making Changes

1.  **Create a Branch:** Before making changes, create a new branch from the `main` branch:
    ```bash
    git checkout main
    git pull origin main # Ensure your main is up-to-date
    git checkout -b feature/your-feature-name # Example: feature/add-nodejs-support
    # or
    git checkout -b fix/issue-description # Example: fix/improve-plan-error-msg
    ```
2.  **Implement:** Make your code changes. Keep changes focused on a single feature or bug fix per branch.
3.  **Follow Style:** Adhere to standard Go formatting (`gofmt`) and linting practices. You can use `contextvibes format` and `contextvibes quality` to help with this. Use `go vet ./...` to catch common issues.
4.  **Test:**
    *   **Manual:** Run the commands you've modified in relevant test projects (e.g., a simple Git repo, a Terraform project, a Go project) to ensure they behave as expected. Use the new `contextvibes test` command for running automated project tests if applicable.
    *   **Automated:** If adding new functions, especially in `internal/`, please add corresponding unit tests (`_test.go` files). Contributions to increase overall test coverage are highly encouraged. Run Go unit tests using:
        ```bash
        go test ./...
        ```
5.  **Commit:** Commit your changes using clear and descriptive commit messages. Consider following the [Conventional Commits](https://www.conventionalcommits.org/) specification (e.g., `feat: ...`, `fix: ...`, `refactor: ...`, `docs: ...`).
    ```bash
    git add .
    git commit -m "feat(plan): Add detection for Rust Cargo.toml"
    ```

## Submitting a Pull Request

1.  **Push:** Push your feature or fix branch to your fork on GitHub:
    ```bash
    git push origin feature/your-feature-name
    ```
2.  **Open PR:** Go to the original `contextvibes/cli` repository on GitHub. GitHub should automatically detect your pushed branch and prompt you to create a Pull Request.
3.  **Describe:** Fill out the Pull Request template, clearly describing the problem you're solving and the changes you've made. Link to any relevant issues.
4.  **Review:** Respond to any feedback or code review comments. The maintainers will review your PR and merge it if it meets the project's standards.

## Finding Ways to Contribute

For a detailed list of known bugs, planned refactorings, potential enhancements, and other ideas for contributions, please see our [ROADMAP.md](ROADMAP.md). We welcome contributions to items listed there or your own suggestions that align with the project's goals!

Thank you for contributing to Context Vibes!

======== END FILE: CONTRIBUTING.md ========

======== FILE: README.md ========

# Context Vibes CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/contextvibes/cli)](https://goreportcard.com/report/github.com/contextvibes/cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<!-- Open in Firebase Studio Button -->
<a href="https://studio.firebase.google.com/import?url=https%3A%2F%2Fgithub.com%2Fcontextvibes%2Fcli"> <!-- Verify this URL! -->
  <picture>
    <source
      media="(prefers-color-scheme: dark)"
      srcset="https://cdn.firebasestudio.dev/btn/open_dark_32.svg">
    <source
      media="(prefers-color-scheme: light)"
      srcset="https://cdn.firebasestudio.dev/btn/open_light_32.svg">
    <img
      height="32"
      alt="Open in Firebase Studio"
      src="https://cdn.firebasestudio.dev/btn/open_blue_32.svg">
  </picture>
</a>
<!-- End Button -->

Context Vibes is a command-line tool designed to streamline common development tasks and generate context for AI assistants. It provides consistent wrappers for Git workflows, Infrastructure as Code (IaC) operations, code quality checks, formatting, testing, and programmatic code modifications, focusing on clear, structured terminal output and detailed background logging.

## Why Context Vibes?

*   **Consistency:** Provides a unified interface and terminal output style for frequent actions (`commit`, `sync`, `deploy`, etc.).
*   **Automation:** Simplifies multi-step processes and provides non-interactive options via the global `--yes` flag. Designed for use in scripts or by AI agents.
*   **AI Integration:**
    *   Generates a `contextvibes.md` context file (`describe`, `diff`) suitable for AI prompts.
    *   Produces structured terminal output suitable for human review or direct AI parsing.
    *   Generates a detailed JSON trace log (default: `contextvibes_ai_trace.log`, configurable) for deeper AI analysis or debugging.
*   **Clarity & Safety:** Uses distinct output formats and requires confirmation for state-changing operations (unless `--yes` is specified).
*   **Configurability:** Supports a `.contextvibes.yaml` file for customizing default behaviors. See the [Configuration Reference](docs/CONFIGURATION_REFERENCE.md) for details.

## Key Features

*   **AI Context Generation:** `describe`, `diff`
*   **Git Workflow Automation:** `kickoff`, `commit`, `sync`, `wrapup`, `status` (Configurable branch/commit rules)
*   **Infrastructure as Code (IaC) Wrappers:** `plan`, `deploy`, `init`
*   **Code Quality & Formatting:** `quality`, `format`
*   **Project Testing & Versioning:** `test`, `version`
*   **Code Modification:** `codemod`

*(For detailed information on each command, see the [Command Reference](docs/COMMAND_REFERENCE.md).)*

## Installation

Ensure you have Go (`1.24` or later recommended) and Git installed.

1.  **Install using `go install`:**
    ```bash
    go install github.com/contextvibes/cli/cmd/contextvibes@latest
    ```
    Installs to `$GOPATH/bin` (usually `$HOME/go/bin`).

2.  **Ensure Installation Directory is in your `PATH`:**
    ```bash
    # Add one of these to your shell profile (.bashrc, .zshrc, etc.)
    export PATH=$(go env GOPATH)/bin:$PATH
    # Or: export PATH=$HOME/go/bin:$PATH
    ```
    Restart your shell or source the profile (`source ~/.bashrc`).

**(Alternative) Installation via Releases:** Download from [GitHub Releases](https://github.com/contextvibes/cli/releases) (*Adjust URL when releases are available*), make executable (`chmod +x`), move to a directory in your `PATH`.

**Dependencies:** Relies on external tools being in your `PATH`: `git`, and potentially `terraform`, `pulumi`, `tflint`, `isort`, `black`, `flake8`, `python`.

## Usage

```bash
contextvibes [command] --help  # General help or help for a specific command
contextvibes [command] [flags] # Run a command
```

**Common Flags:**

*   `-y`, `--yes`: Assume 'yes' to all confirmation prompts.
*   `--ai-log-file <path>`: Specify a path for the detailed AI JSON log.
*   `--log-level-ai <level>`: Set the minimum level for the AI log file (debug, info, warn, error).

*(See the [Command Reference](docs/COMMAND_REFERENCE.md) for all commands and flags.)*

**Examples:**

```bash
# Start a new feature branch (prompts for name if needed)
contextvibes kickoff --branch feature/add-user-auth

# Describe the project for an AI (prompts for task description)
contextvibes describe -o my_context.md

# Apply code formatting
contextvibes format

# Check code quality
contextvibes quality

# Run project tests (e.g., for a Go project, passing -v flag)
contextvibes test -v

# Commit work (message required, interactive confirmation)
contextvibes commit -m "feat(auth): Implement OTP login"

# Sync non-interactively
contextvibes sync -y

# Display CLI version
contextvibes version

# Apply programmatic changes from a script
contextvibes codemod --script ./changes.json```

## Documentation

*   **[Overview & Installation](README.md):** (This file) High-level features and setup.
*   **[Command Reference](docs/COMMAND_REFERENCE.md):** Detailed syntax, flags, examples, and exit codes for every command. **Use this for specific command usage.**
*   **[Configuration Reference](docs/CONFIGURATION_REFERENCE.md):** Full details on configuring the CLI via `.contextvibes.yaml`. **Use this to customize behavior.**
*   **[Contributing Guidelines](CONTRIBUTING.md):** How to contribute code, report issues, and set up a development environment.
*   **[Changelog](CHANGELOG.md):** History of notable changes in each release.
*   **[Roadmap](ROADMAP.md):** Future plans and development direction.

*(Additional Tutorials and How-To Guides may be added to the `docs/` directory.)*

## Important: Ignoring Generated Files

It is strongly recommended to add generated files like `contextvibes.md`, `contextvibes_ai_trace.log`, `*.log`, and `tfplan.out` to your project's `.gitignore` file.

## Terminal Output vs. AI Log File

Context Vibes uses two distinct output mechanisms:

1.  **Terminal Output (stdout/stderr):** For human readability and high-level status/errors. Uses structured prefixes (`SUMMARY:`, `INFO:`, etc.).
2.  **AI Log File (JSON):** Written to `contextvibes_ai_trace.log` by default (configurable). Contains a detailed, structured trace for AI analysis or debugging.

## Code of Conduct

Act professionally and respectfully. Be kind, considerate, and welcoming. Harassment or exclusionary behavior will not be tolerated.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

======== END FILE: README.md ========

======== FILE: ROADMAP.md ========

# Context Vibes CLI - Roadmap

This document outlines the planned features and future direction for the Context Vibes CLI. Priorities and features may change based on feedback and community contributions. This roadmap aims to guide development efforts and inform potential contributors about areas where help is welcome.

## Near Term (Next 1-2 Minor Releases)

*   [ ] **Comprehensive Unit Testing:** *(High Priority)*
    *   Expand unit test coverage significantly, focusing on:
        *   `internal/git.GitClient` methods.
        *   `internal/exec` package components.
        *   `internal/config` package logic.
        *   `internal/tools` utilities.
        *   `internal/project` package.
    *   Goal: Ensure robustness, facilitate safe refactoring, and improve maintainability.
*   [ ] **Implement Configurable Commit Message Validation in `cmd/commit.go`:**
    *   Ensure the `commit` command fully utilizes the `validation.commitMessage.enable` and `validation.commitMessage.pattern` settings from `.contextvibes.yaml`, aligning its behavior with the README and configuration capabilities.
*   [ ] **Refactor `describe` Command:**
    *   Break down the large `RunE` function in `cmd/describe.go` into smaller, more manageable, and testable helper functions within the `cmd` package. This will improve readability and maintainability.
*   [ ] **Enhance `quality` for `go mod tidy`:**
    *   Modify the `go mod tidy` check in the `quality` command to detect if `go.mod` or `go.sum` files *were modified* by `go mod tidy`. If modifications occur, the check should fail (similar to how `go fmt` non-compliance is handled), prompting the user to commit the changes.
*   [ ] **`init` Command Flag Enhancements:**
    *   Add flags to the `init` command (especially for Terraform) to pass common options to the underlying initialization tools (e.g., `terraform init -upgrade`, `terraform init -reconfigure`).

## Medium Term

*   [ ] **Context File Management Strategy (`contextvibes.md` "Memory"):**
    *   Explore and implement strategies for managing the `contextvibes.md` file, particularly given that the `diff` command currently overwrites it.
    *   Options to consider:
        *   An optional backup flag for `diff` before overwriting.
        *   Allowing timestamped or uniquely named output files for `diff` and `describe`.
        *   Alternative approaches to providing diff context alongside full project context.
*   [ ] **Global Verbose Flag:**
    *   Introduce a global `--verbose` or `-v` flag. This would enable more detailed output during command execution, such as showing the full commands being run by the `ExecutorClient` or more granular internal step logging to the terminal.
*   [ ] **Expand Quality Tool Integrations:**
    *   Investigate and add support for more widely-used linters and static analysis tools.
    *   Examples:
        *   Go: `golangci-lint`.
        *   Python: `mypy` (for type checking).
*   [ ] **Richer `describe` Project Structure Analysis:**
    *   Enhance the `describe` command to provide more intelligent and detailed project structure information beyond simple `tree` or `ls` output.
    *   Examples: Analyzing `go.mod` for Go project dependencies/packages, `pyproject.toml`/`requirements.txt` for Python, `package.json` for Node.js (if/when supported).
*   [ ] **Standalone `update` Command (Revisit):**
    *   Consider re-introducing a simpler `update` command (e.g., `git pull --rebase` on the current branch with confirmation) for users who prefer an explicit update action without the push component of the `sync` command.
*   [ ] **Git Stash Integration (Optional):**
    *   Explore adding a `contextvibes stash` command or integrating optional stashing capabilities into commands like `kickoff` or `sync` to more gracefully handle non-clean working directories, prompting the user to stash/unstash.
*   [ ] **Enhanced `.idx/airules.md` Interaction (Re-evaluate Scope):**
    *   Currently, `contextvibes` reads `.idx/airules.md` for context in `describe`.
    *   Re-evaluate if more active generation, validation, or maintenance features for this specific IDE context file are desired by users or align with the CLI's core purpose. This might involve interactive creation or specific structural validation if it becomes a widely adopted standard.

## Long Term / Vision

*   [ ] **Broader Project Type Support:**
    *   Incrementally expand project detection and relevant command adaptations (`quality`, `format`, `test`, `init`, `plan`, `deploy` steps) for other common development ecosystems.
    *   Examples: Node.js (npm/yarn commands, linters), Java (Maven/Gradle tasks), Rust (Cargo tasks), Docker (Dockerfile linting).
*   [ ] **Plugin System or Extensibility:**
    *   Explore architectural changes to allow users or teams to more easily extend `contextvibes` with custom commands, project-type specific logic, or quality tool integrations without modifying the core CLI codebase.
*   [ ] **CI/CD Integration Enhancements:**
    *   Provide clearer examples, documentation, or specific flags/output modes (e.g., machine-readable output for some commands) to facilitate the robust use of `contextvibes` commands within automated CI/CD pipelines.
*   [ ] **Configuration File Validation:**
    *   Add a dedicated command (e.g., `contextvibes config validate`) to check the syntax and potentially the semantic correctness of the `.contextvibes.yaml` file.
*   [ ] **Improved AI Prompting & Context (`contextvibes.md`):**
    *   Continuously refine the structure, content, and verbosity of the `contextvibes.md` file based on evolving best practices for prompting various AI models and user feedback.

---

This roadmap is a living document. Feedback, suggestions, and contributions are highly welcome! Please open an issue on the GitHub repository to discuss roadmap items or propose new features.

======== END FILE: ROADMAP.md ========

======== FILE: cmd/codemod.go ========

// cmd/codemod.go
package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/contextvibes/cli/internal/codemod" // Using the new types package
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

const defaultCodemodFilename = "codemod.json" // Default script filename

var codemodScriptPath string

var codemodCmd = &cobra.Command{
	Use:   "codemod [--script <file.json>]",
	Short: "Applies programmatic code modifications or deletions from a JSON script.",
	Long: `Reads a JSON script file describing a series of operations to be applied to
specified files in the codebase. This enables automated or AI-assisted refactoring and cleanup.

If --script is not provided, it looks for '` + defaultCodemodFilename + `' in the current directory.

The JSON script should be an array of objects, where each object defines:
  - "file_path": The path to the file to be modified or deleted.
  - "operations": An array of operation objects for that file.

Currently supported operation types:
  - "regex_replace": Performs find/replace on file content.
    Required fields: "type": "regex_replace", "find_regex": "...", "replace_with": "..."
  - "delete_file": Deletes the specified file_path.
    Required fields: "type": "delete_file"

Requires confirmation before writing/deleting, unless --yes is specified.`,
	Example: `  contextvibes codemod # Looks for codemod.json
  contextvibes codemod --script ./my_refactor_script.json
  contextvibes codemod -s ./cleanup.json -y`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)

		scriptToLoad := codemodScriptPath

		if scriptToLoad == "" {
			presenter.Info("No --script provided, attempting to load default: %s", defaultCodemodFilename)
			if _, err := os.Stat(defaultCodemodFilename); os.IsNotExist(err) {
				presenter.Error("Default codemod script '%s' not found and no --script flag provided.", defaultCodemodFilename)
				presenter.Advice("Create '%s' in the current directory or use the --script flag to specify a file.", defaultCodemodFilename)
				return fmt.Errorf("no codemod script specified or found")
			}
			scriptToLoad = defaultCodemodFilename
		}

		presenter.Summary("Applying codemod script: %s", scriptToLoad)

		scriptData, err := os.ReadFile(scriptToLoad)
		if err != nil {
			presenter.Error("Failed to read codemod script file '%s': %v", scriptToLoad, err)
			logger.Error("codemod: failed to read script file", slog.String("path", scriptToLoad), slog.Any("error", err))
			return err
		}

		var script codemod.ChangeScript
		if err := json.Unmarshal(scriptData, &script); err != nil {
			presenter.Error("Failed to parse codemod script JSON from '%s': %v", scriptToLoad, err)
			logger.Error("codemod: failed to parse script json", slog.String("path", scriptToLoad), slog.Any("error", err))
			return err
		}

		if len(script) == 0 {
			presenter.Info("Codemod script is empty. No changes to apply.")
			return nil
		}

		totalFilesModified := 0
		totalFilesDeleted := 0
		totalOperationsAttempted := 0
		totalOperationsSucceeded := 0

		for _, fileChangeSet := range script {
			presenter.Header("Processing target: %s", fileChangeSet.FilePath)

			onlyDelete := len(fileChangeSet.Operations) == 1 && fileChangeSet.Operations[0].Type == "delete_file"
			fileExists := false
			fileInfo, statErr := os.Stat(fileChangeSet.FilePath)
			if statErr == nil {
				fileExists = true
			} else if !os.IsNotExist(statErr) {
				presenter.Error("Error checking file %s: %v. Skipping.", fileChangeSet.FilePath, statErr)
				logger.Error("codemod: failed to stat file", slog.String("path", fileChangeSet.FilePath), slog.Any("error", statErr))
				continue
			}

			if !fileExists && !onlyDelete {
				presenter.Error("File not found: %s. Skipping operations (except delete_file).", fileChangeSet.FilePath)
				logger.Error("codemod: file not found, skipping changeset", slog.String("path", fileChangeSet.FilePath))
				continue
			}

			var currentFileContent string
			var contentBeforeOpsForThisFile string
			if fileExists && !onlyDelete {
				fileContentBytes, readErr := os.ReadFile(fileChangeSet.FilePath)
				if readErr != nil {
					presenter.Error("Failed to read file %s: %v. Skipping.", fileChangeSet.FilePath, readErr)
					logger.Error("codemod: failed to read file for modification", slog.String("path", fileChangeSet.FilePath), slog.Any("error", readErr))
					continue
				}
				currentFileContent = string(fileContentBytes)
				contentBeforeOpsForThisFile = currentFileContent
			} else {
				contentBeforeOpsForThisFile = ""
				currentFileContent = ""
			}

			fileWasDeleted := false

		operationsLoop: // Label for the operations loop for this file
			for opIndex, op := range fileChangeSet.Operations {
				totalOperationsAttempted++
				presenter.Step("Attempting operation %d: type='%s', desc='%s'", opIndex+1, op.Type, op.Description)

				opSucceeded := false
				contentBeforeThisOp := currentFileContent

				switch op.Type {
				case "regex_replace":
					if fileWasDeleted {
						presenter.Warning("Skipping regex_replace on '%s' as file was already deleted by a previous operation.", fileChangeSet.FilePath)
						continue // to next operation
					}
					if !fileExists { // This check might be redundant if fileWasDeleted is handled correctly
						presenter.Error("Cannot apply regex_replace: file '%s' does not exist.", fileChangeSet.FilePath)
						continue // to next operation
					}
					if op.FindRegex == "" {
						presenter.Warning("Skipping regex_replace for '%s': find_regex is empty.", fileChangeSet.FilePath)
						logger.Warn("codemod: regex_replace skipped, empty find_regex", slog.String("file", fileChangeSet.FilePath))
						continue // to next operation
					}
					re, compileErr := regexp.Compile(op.FindRegex)
					if compileErr != nil {
						presenter.Error("Invalid regex '%s' for file '%s': %v. Skipping operation.", op.FindRegex, fileChangeSet.FilePath, compileErr)
						logger.Error("codemod: invalid regex in script", slog.String("file", fileChangeSet.FilePath), slog.String("regex", op.FindRegex), slog.Any("error", compileErr))
						continue // to next operation
					}

					currentFileContent = re.ReplaceAllString(currentFileContent, op.ReplaceWith)
					if currentFileContent != contentBeforeThisOp {
						opSucceeded = true
						presenter.Info("  Applied regex_replace to '%s'.", fileChangeSet.FilePath)
						logger.Info("codemod: applied regex_replace", slog.String("file", fileChangeSet.FilePath), slog.String("find", op.FindRegex), slog.String("replace", op.ReplaceWith))
					} else {
						opSucceeded = true
						presenter.Info("  Regex_replace on '%s' resulted in no changes.", fileChangeSet.FilePath)
						logger.Info("codemod: regex_replace no change", slog.String("file", fileChangeSet.FilePath), slog.String("find", op.FindRegex))
					}

				case "delete_file":
					if !fileExists { // If file didn't exist at the start of processing this FileChangeSet
						presenter.Info("File '%s' already does not exist. 'delete_file' operation considered successful.", fileChangeSet.FilePath)
						opSucceeded = true
						fileWasDeleted = true // Mark as "conceptually" deleted for this changeset
						break operationsLoop  // Exit the operations loop for this file
					}
					// If fileWasDeleted is true, it means a previous "delete_file" op in *this same FileChangeSet* already deleted it.
					if fileWasDeleted {
						presenter.Info("File '%s' already actioned for deletion by a previous operation in this set.", fileChangeSet.FilePath)
						opSucceeded = true   // Considered success as the goal is achieved
						break operationsLoop // Exit the operations loop for this file
					}

					presenter.Info("Operation requests deletion of file: %s", fileChangeSet.FilePath)
					deleteConfirmed := false
					if assumeYes {
						presenter.Info("Deleting file '%s' (confirmation bypassed via --yes).", fileChangeSet.FilePath)
						deleteConfirmed = true
					} else {
						var promptErr error
						deleteConfirmed, promptErr = presenter.PromptForConfirmation(fmt.Sprintf("Permanently delete file '%s'?", fileChangeSet.FilePath))
						if promptErr != nil {
							presenter.Error("Error during delete confirmation for '%s': %v. Skipping deletion.", fileChangeSet.FilePath, promptErr)
							logger.Error("codemod: delete confirmation error", slog.String("file", fileChangeSet.FilePath), slog.Any("error", promptErr))
							continue // to next operation
						}
					}

					if deleteConfirmed {
						err := os.Remove(fileChangeSet.FilePath)
						if err != nil {
							if os.IsNotExist(err) { // File was deleted by another process between Stat and Remove
								presenter.Warning("File '%s' was not found during deletion attempt (possibly deleted externally).", fileChangeSet.FilePath)
								opSucceeded = true    // Goal achieved
								fileWasDeleted = true // Mark as deleted
							} else {
								presenter.Error("Failed to delete file '%s': %v", fileChangeSet.FilePath, err)
								logger.Error("codemod: failed to delete file", slog.String("path", fileChangeSet.FilePath), slog.Any("error", err))
								// opSucceeded remains false
							}
						} else {
							presenter.Success("Successfully deleted file: %s", fileChangeSet.FilePath)
							logger.Info("codemod: deleted file", slog.String("path", fileChangeSet.FilePath))
							opSucceeded = true
							fileWasDeleted = true
							totalFilesDeleted++
						}
					} else {
						presenter.Info("Skipped deletion of file '%s' by user.", fileChangeSet.FilePath)
						logger.Info("codemod: delete skipped by user", slog.String("file", fileChangeSet.FilePath))
						// opSucceeded remains false
					}

					if fileWasDeleted {
						break operationsLoop // Critical: exit the operations loop for this file after delete
					}

				default:
					presenter.Warning("Unsupported operation type: '%s' for file '%s'. Skipping.", op.Type, fileChangeSet.FilePath)
					logger.Warn("codemod: unsupported operation type", slog.String("type", op.Type), slog.String("file", fileChangeSet.FilePath))
				} // end switch op.Type

				if opSucceeded {
					totalOperationsSucceeded++
				}
			} // end operationsLoop

			// If file wasn't deleted by an operation in this set, check if content changed and needs writing
			if !fileWasDeleted && currentFileContent != contentBeforeOpsForThisFile {
				presenter.Info("File '%s' has pending modifications.", fileChangeSet.FilePath)

				confirmedWrite := false
				if assumeYes {
					presenter.Info("Writing changes to '%s' (confirmation bypassed via --yes).", fileChangeSet.FilePath)
					confirmedWrite = true
				} else {
					var promptErr error
					confirmedWrite, promptErr = presenter.PromptForConfirmation(fmt.Sprintf("Write modified content to '%s'?", fileChangeSet.FilePath))
					if promptErr != nil {
						presenter.Error("Error during write confirmation for '%s': %v. Skipping write.", fileChangeSet.FilePath, promptErr)
						logger.Error("codemod: write confirmation error", slog.String("file", fileChangeSet.FilePath), slog.Any("error", promptErr))
						continue // to next file in the script
					}
				}

				if confirmedWrite {
					var perm os.FileMode = 0644
					if fileInfo != nil { // Use the FileInfo from the initial Stat if file existed
						perm = fileInfo.Mode().Perm()
					} else {
						// This case should be rare if we are writing, as it implies file didn't exist but was modified.
						logger.Warn("codemod: could not get original file permissions for '%s' (was it created by an operation?), using default 0644", slog.String("file", fileChangeSet.FilePath))
					}

					err := os.WriteFile(fileChangeSet.FilePath, []byte(currentFileContent), perm)
					if err != nil {
						presenter.Error("Failed to write changes to %s: %v", fileChangeSet.FilePath, err)
						logger.Error("codemod: failed to write file", slog.String("path", fileChangeSet.FilePath), slog.Any("error", err))
					} else {
						presenter.Success("Successfully updated %s", fileChangeSet.FilePath)
						totalFilesModified++
					}
				} else {
					presenter.Info("Skipped writing changes to %s due to user cancellation.", fileChangeSet.FilePath)
					logger.Info("codemod: write skipped by user", slog.String("file", fileChangeSet.FilePath))
				}
			} else if !fileWasDeleted { // Only print this if not deleted AND not modified from its original state for this FileChangeSet
				presenter.Info("No effective changes made to %s after all operations.", fileChangeSet.FilePath)
			}
			presenter.Newline()
		} // end fileChangeSet loop

		presenter.Separator()
		presenter.Summary("Codemod script execution finished.")
		presenter.Detail("Files Modified: %d", totalFilesModified)
		presenter.Detail("Files Deleted: %d", totalFilesDeleted)
		presenter.Detail("Operations Attempted: %d", totalOperationsAttempted)
		presenter.Detail("Operations Succeeded (may include no-ops/already deleted): %d", totalOperationsSucceeded)
		return nil
	},
}

func init() {
	codemodCmd.Flags().StringVarP(&codemodScriptPath, "script", "s", "", "Path to the JSON codemod script file (default: "+defaultCodemodFilename+")")
	rootCmd.AddCommand(codemodCmd)
}

======== END FILE: cmd/codemod.go ========

======== FILE: cmd/commit.go ========

// cmd/commit.go

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp" // Added for commit message validation
	"strings"

	"github.com/contextvibes/cli/internal/config" // Import for DefaultCommitMessagePattern
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui" // Import the presenter package
	"github.com/spf13/cobra"
)

var commitMessageFlag string

// The hardcoded conventionalCommitRegexPattern is no longer the primary source.
// It will be determined by configuration or fallback to config.DefaultCommitMessagePattern.
// const conventionalCommitRegexPattern = `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_]+\))?:\s.+`

var commitCmd = &cobra.Command{
	Use:   "commit -m <message>",
	Short: "Stages all changes and commits locally with a provided message.",
	Long: `Stages all current changes (tracked and untracked) in the working directory
and commits them locally using the message provided via the -m/--message flag.

Commit message validation is active by default, expecting a Conventional Commits format.
This can be configured (pattern or disabled) in '.contextvibes.yaml'.
Default pattern if validation is enabled and no custom pattern is set:
  ` + config.DefaultCommitMessagePattern + `
Example (using default Conventional Commits):
  feat(login): add forgot password button
  fix(api): correct user data validation

Requires confirmation before committing unless -y/--yes is specified.
Does NOT automatically push.`,
	Example: `  contextvibes commit -m "feat(auth): Implement OTP login"
  contextvibes commit -m "fix: Correct typo in user model" -y
  contextvibes commit -m "My custom message" # (if validation is disabled or pattern allows)
  contextvibes commit --config-validation-pattern="^TASK-[0-9]+: .+" -m "TASK-123: Implement feature" # (Example of custom pattern if it were a flag, actual via .contextvibes.yaml)`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			fmt.Fprintln(os.Stderr, "[ERROR] Internal error: logger not initialized")
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		if LoadedAppConfig == nil {
			presenter.Error("Internal error: Application configuration not loaded.")
			logger.ErrorContext(ctx, "Commit failed: LoadedAppConfig is nil", slog.String("source_command", "commit"))
			return errors.New("application configuration not loaded")
		}

		// --- Validate Input ---
		if strings.TrimSpace(commitMessageFlag) == "" {
			errMsgForUser := "Commit message is required. Please provide one using the -m flag."
			errMsgForError := "commit message is required via -m flag"
			presenter.Error(errMsgForUser)
			presenter.Advice("Example: `%s commit -m \"feat(module): Your message\"`", cmd.CommandPath())
			logger.ErrorContext(ctx, "Commit failed: missing required message flag", slog.String("source_command", "commit"))
			return errors.New(errMsgForError)
		}
		finalCommitMessage := commitMessageFlag

		// --- Validate Commit Message Format Based on Configuration ---
		commitMsgValidationRule := LoadedAppConfig.Validation.CommitMessage
		validationIsEnabled := commitMsgValidationRule.Enable == nil || *commitMsgValidationRule.Enable // Default to true if nil

		effectivePattern := ""
		patternSource := ""

		if validationIsEnabled {
			logger.InfoContext(ctx, "Commit message validation is enabled.", slog.String("source_command", "commit"))
			effectivePattern = commitMsgValidationRule.Pattern
			patternSource = "from .contextvibes.yaml"

			if effectivePattern == "" {
				effectivePattern = config.DefaultCommitMessagePattern // Fallback to built-in default
				patternSource = "default built-in"
				logger.DebugContext(ctx, "Using default built-in commit message pattern because configured pattern is empty.", slog.String("source_command", "commit"), slog.String("pattern", effectivePattern))
			} else {
				logger.DebugContext(ctx, "Using commit message pattern from configuration.", slog.String("source_command", "commit"), slog.String("pattern", effectivePattern))
			}

			if effectivePattern == "" {
				// This should not happen if DefaultCommitMessagePattern is always defined
				presenter.Error("Internal Error: Commit message validation is enabled but no pattern is defined or defaulted.")
				logger.ErrorContext(ctx, "Commit failed: validation enabled but no pattern available", slog.String("source_command", "commit"))
				return errors.New("commit validation pattern misconfiguration (empty effective pattern)")
			}

			commitMsgRe, compileErr := regexp.Compile(effectivePattern)
			if compileErr != nil {
				errMsgForUser := fmt.Sprintf("Internal error: Invalid commit message validation pattern ('%s') from %s.", effectivePattern, patternSource)
				errMsgForError := "invalid commit message validation regex"
				presenter.Error(errMsgForUser)
				presenter.Advice("Error details: %v", compileErr)
				presenter.Advice("Please check your .contextvibes.yaml or report this issue if using the default pattern.")
				logger.ErrorContext(ctx, "Commit failed: invalid regex for commit message",
					slog.String("source_command", "commit"),
					slog.String("pattern", effectivePattern),
					slog.String("pattern_source", patternSource),
					slog.String("error", compileErr.Error()))
				return errors.New(errMsgForError)
			}

			if !commitMsgRe.MatchString(finalCommitMessage) {
				errMsgForUser := "Invalid commit message format."
				errMsgForError := "invalid commit message format"
				presenter.Error(errMsgForUser)
				presenter.Advice("Message should match the pattern (%s): `%s`", patternSource, effectivePattern)
				if patternSource != "default built-in" && effectivePattern == config.DefaultCommitMessagePattern {
					presenter.Detail(" (Note: Configured pattern seems to be the same as the default Conventional Commits pattern.)")
				} else if effectivePattern == config.DefaultCommitMessagePattern {
					presenter.Detail("  Default pattern expects: <type>(<scope>): <subject>")
					presenter.Detail("  Valid types: BREAKING, feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert")
					presenter.Detail("  Example: feat(ui): add new save button")
				}
				presenter.Detail("  Your message: \"%s\"", finalCommitMessage)
				logger.ErrorContext(ctx, "Commit failed: invalid message format",
					slog.String("source_command", "commit"),
					slog.String("commit_message", finalCommitMessage),
					slog.String("pattern", effectivePattern),
					slog.String("pattern_source", patternSource))
				return errors.New(errMsgForError)
			}
			logger.DebugContext(ctx, "Commit message format validated successfully",
				slog.String("source_command", "commit"),
				slog.String("message", finalCommitMessage),
				slog.String("pattern", effectivePattern),
				slog.String("pattern_source", patternSource))
		} else {
			presenter.Info("Commit message validation is disabled by configuration (.contextvibes.yaml).")
			logger.InfoContext(ctx, "Commit message validation skipped due to configuration", slog.String("source_command", "commit"))
		}

		// --- Summary ---
		presenter.Summary("Attempting to stage and commit changes locally.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get working directory: %v", err)
			logger.ErrorContext(ctx, "Commit: Failed getwd", slog.String("error", err.Error()))
			return err
		}
		// Pass relevant config values to GitClientConfig
		gitCfg := git.GitClientConfig{
			Logger:                logger,
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch,
		}
		logger.DebugContext(ctx, "Initializing GitClient with effective app config",
			slog.String("source_command", "commit"),
			slog.String("remote", gitCfg.DefaultRemoteName),
			slog.String("mainBranch", gitCfg.DefaultMainBranchName))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed to initialize Git client: %v", err)
			return err // Client logs details
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "commit"))

		// --- Pre-Checks ---
		currentBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			presenter.Error("Failed to get current branch name: %v", err)
			return err
		}
		// Use client.MainBranchName() which gets the effective main branch from its config
		if currentBranch == client.MainBranchName() {
			errMsg := fmt.Sprintf("Cannot commit directly on the main branch ('%s').", client.MainBranchName())
			presenter.Error(errMsg)
			presenter.Advice("Use `contextvibes kickoff` to start a new feature/fix branch first.")
			logger.ErrorContext(ctx, "Commit failed: attempt to commit on main branch", slog.String("source_command", "commit"), slog.String("branch", currentBranch))
			return errors.New("commit on main branch is disallowed by this command")
		}

		// --- Stage & Check ---
		logger.DebugContext(ctx, "Attempting to stage all changes (git add .)", slog.String("source_command", "commit"))
		if err := client.AddAll(ctx); err != nil {
			presenter.Error("Failed to stage changes: %v", err)
			return err // Client logs details
		}
		logger.DebugContext(ctx, "Staging completed.", slog.String("source_command", "commit"))

		hasStaged, err := client.HasStagedChanges(ctx)
		if err != nil {
			presenter.Error("Failed to check for staged changes: %v", err)
			return err // Client logs details
		}
		if !hasStaged {
			presenter.Info("No changes were staged for commit (working directory may have been clean or `git add .` staged nothing).")
			logger.InfoContext(ctx, "No staged changes found to commit.", slog.String("source_command", "commit"))
			return nil
		}

		// --- Fetch Git Status Details for Display ---
		statusOutput, _, statusErr := client.GetStatusShort(ctx)
		if statusErr != nil {
			logger.WarnContext(ctx, "Could not get short status for info block, proceeding with commit attempt.", slog.String("source_command", "commit"), slog.String("error", statusErr.Error()))
		}

		// --- Consolidated INFO Block ---
		presenter.Newline()
		presenter.InfoPrefixOnly()

		fmt.Fprintf(presenter.Out(), "  Branch: %s\n", currentBranch)
		fmt.Fprintf(presenter.Out(), "  Commit Message:\n    \"%s\"\n", finalCommitMessage)
		if validationIsEnabled {
			fmt.Fprintf(presenter.Out(), "  Validation Pattern (%s):\n    `%s`\n", patternSource, effectivePattern)
		} else {
			fmt.Fprintln(presenter.Out(), "  Validation: Disabled by configuration")
		}
		fmt.Fprintf(presenter.Out(), "  Staged Changes:\n")

		if statusErr != nil {
			fmt.Fprintf(presenter.Out(), "    (Could not retrieve status details for display: %v)\n", statusErr)
		} else if strings.TrimSpace(statusOutput) == "" {
			fmt.Fprintln(presenter.Out(), "    (Staged changes detected, but `git status --short` was unexpectedly empty)")
		} else {
			scanner := bufio.NewScanner(strings.NewReader(statusOutput))
			for scanner.Scan() {
				fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
		}
		presenter.Newline()

		// --- Confirmation ---
		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(ctx, "Confirmation bypassed via --yes flag", slog.String("source_command", "commit"))
			confirmed = true
		} else {
			var promptErr error
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with this commit?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading confirmation for commit", slog.String("source_command", "commit"), slog.String("error", promptErr.Error()))
				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Commit aborted by user.")
			logger.InfoContext(ctx, "Commit aborted by user confirmation.", slog.String("source_command", "commit"))
			return nil
		}
		logger.DebugContext(ctx, "Proceeding with commit after confirmation.", slog.String("source_command", "commit"))

		// --- Execute Commit ---
		presenter.Info("Executing commit...")
		logger.DebugContext(ctx, "Attempting commit via GitClient.Commit", slog.String("source_command", "commit"))
		if err := client.Commit(ctx, finalCommitMessage); err != nil {
			presenter.Error("Commit command failed: %v", err)
			logger.ErrorContext(ctx, "client.Commit method failed", slog.String("source_command", "commit"), slog.String("error", err.Error()))
			return err
		}

		// --- Success & Advice ---
		presenter.Newline()
		presenter.Success("Commit created successfully locally.")
		presenter.Advice("Consider syncing your changes using `contextvibes sync`.")
		logger.InfoContext(ctx, "Commit successful", slog.String("source_command", "commit"), slog.String("commit_message", finalCommitMessage))
		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessageFlag, "message", "m", "", "Commit message (required)")
	// Long description of commitCmd already describes the default pattern and configurability
	rootCmd.AddCommand(commitCmd)
}

======== END FILE: cmd/commit.go ========

======== FILE: cmd/contextvibes/main.go ========

package main

import (
	"github.com/contextvibes/cli/cmd"
)

func main() {
	cmd.Execute()
}

======== END FILE: cmd/contextvibes/main.go ========

======== FILE: cmd/deploy.go ========

// cmd/deploy.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/project"
	// "github.com/contextvibes/cli/internal/tools" // Only needed for os.Stat now, if CommandExists/ExecuteCommand are removed
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
	// Use the execClientInterface, assuming it's defined perhaps in a shared cmd/helpers.go or locally
	// For simplicity, let's assume it's accessible or redefine it if needed. We'll use the global ExecClient for now.
	// No direct import of internal/exec needed if using global ExecClient
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys infrastructure changes (terraform apply, pulumi up).",
	Long: `Detects the project type (Terraform, Pulumi), explains the deployment action,
and executes the deployment after confirmation (unless -y/--yes is specified).

- Terraform: Requires 'tfplan.out' from 'contextvibes plan'. Runs 'terraform apply tfplan.out'.
- Pulumi: Runs 'pulumi up', which internally includes a preview and confirmation.`,
	Example: `  # For Terraform:
  contextvibes plan    # First, generate the plan file (tfplan.out)
  contextvibes deploy  # Explain plan and prompt to apply tfplan.out
  contextvibes deploy -y # Apply tfplan.out without prompting

  # For Pulumi:
  contextvibes plan    # (Optional) Preview changes first
  contextvibes deploy  # Explain and run 'pulumi up' (includes preview & confirm)
  contextvibes deploy -y # Run 'pulumi up' without contextvibes confirmation`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		// Use global ExecClient from cmd/root.go
		if ExecClient == nil {
			return fmt.Errorf("internal error: executor client not initialized")
		}
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Attempting to deploy infrastructure changes.")

		cwd, err := os.Getwd()
		if err != nil {
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Deploy: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return wrappedErr
		}

		presenter.Info("Detecting project type in %s...", presenter.Highlight(cwd))
		projType, err := project.Detect(cwd)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to detect project type: %w", err)
			logger.ErrorContext(ctx, "Deploy: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return wrappedErr
		}

		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "deploy"), slog.String("type", string(projType)))

		switch projType {
		case project.Terraform:
			// Pass ExecClient to the helper function
			return executeTerraformDeploy(ctx, presenter, logger, ExecClient, cwd, assumeYes)
		case project.Pulumi:
			// Pass ExecClient to the helper function
			return executePulumiDeploy(ctx, presenter, logger, ExecClient, cwd, assumeYes)
		case project.Go:
			presenter.Info("Deploy command is not applicable for Go projects.")
			return nil
		case project.Python:
			presenter.Info("Deploy command is not applicable for Python projects.")
			return nil
		case project.Unknown:
			errMsgForUser := "Unknown project type detected. Cannot determine deploy action."
			errMsgForError := "unknown project type detected"
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "deploy"))
			return errors.New(errMsgForError)
		default:
			errMsgForUser := fmt.Sprintf("Internal error: Unhandled project type '%s'", projType)
			errMsgForError := fmt.Sprintf("internal error: unhandled project type '%s'", projType)
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "deploy"))
			return errors.New(errMsgForError)
		}
	},
}

// Define an interface matching the methods used by the helpers below.
// This makes the helpers testable independently of the global ExecClient.
type execDeployClientInterface interface {
	CommandExists(commandName string) bool
	Execute(ctx context.Context, dir string, commandName string, args ...string) error
}

// executeTerraformDeploy now accepts execClient
func executeTerraformDeploy(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execDeployClientInterface, dir string, skipConfirm bool) error {
	tool := "terraform"
	planFile := "tfplan.out"
	planFilePath := filepath.Join(dir, planFile)
	args := []string{"apply", "-auto-approve", planFile}

	if !execClient.CommandExists(tool) { // Use execClient
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Terraform is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Terraform deploy prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	// Check for plan file using standard os.Stat - this doesn't involve executing a command
	logger.DebugContext(ctx, "Checking for Terraform plan file", slog.String("path", planFilePath))
	if _, err := os.Stat(planFilePath); os.IsNotExist(err) {
		errMsgForUser := fmt.Sprintf("Terraform plan file '%s' not found.", planFile)
		errMsgForError := "terraform plan file not found"
		presenter.Error(errMsgForUser)
		presenter.Advice("Please run `contextvibes plan` first to generate the plan file.")
		logger.Error("Terraform deploy prerequisite failed: plan file missing", slog.String("plan_file", planFile))
		return errors.New(errMsgForError)
	} else if err != nil {
		errMsgForUser := fmt.Sprintf("Error checking for plan file '%s': %v", planFilePath, err)
		errMsgForErrorBase := "error checking for plan file"
		presenter.Error(errMsgForUser)
		logger.Error("Terraform deploy: error stating plan file", slog.String("plan_file", planFilePath), slog.String("error", err.Error()))
		return fmt.Errorf("%s %s: %w", errMsgForErrorBase, planFilePath, err)
	}
	presenter.Info("Using Terraform plan file: %s", presenter.Highlight(planFile))

	// Confirmation logic remains the same
	presenter.Newline()
	presenter.Info("Proposed Deploy Action:")
	presenter.Detail("Apply the Terraform plan '%s' using command:", planFile)
	presenter.Detail("  %s %s", tool, strings.Join(args, " "))
	presenter.Newline()

	confirmed := false
	if skipConfirm {
		presenter.Info("Confirmation prompt bypassed via --yes flag.")
		logger.InfoContext(ctx, "Confirmation bypassed via flag", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Bool("yes_flag", true))
		confirmed = true
	} else {
		var promptErr error
		confirmed, promptErr = presenter.PromptForConfirmation("Proceed with Terraform deployment?")
		if promptErr != nil {
			logger.ErrorContext(ctx, "Error reading deploy confirmation", slog.String("tool", tool), slog.String("error", promptErr.Error()))
			return promptErr
		}
	}

	if !confirmed {
		presenter.Info("Terraform deployment aborted by user.")
		logger.InfoContext(ctx, "Deploy aborted by user confirmation", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Bool("confirmed", false))
		return nil
	}
	logger.DebugContext(ctx, "Proceeding after deploy confirmation", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Bool("confirmed", true))

	// Execution using execClient
	presenter.Newline()
	presenter.Info("Starting Terraform apply...")
	logger.Info("Executing terraform apply", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Any("args", args))

	// Use execClient.Execute - terraform apply pipes its own output
	err := execClient.Execute(ctx, dir, tool, args...) // Use execClient
	if err != nil {
		// Error message from Execute should contain exit code info.
		// User will see the piped output from terraform apply itself.
		errMsgForUser := "'terraform apply' command failed."
		errMsgForError := "terraform apply command failed"
		presenter.Error(errMsgForUser)
		logger.Error("Terraform apply command failed", slog.String("source_command", "deploy"), slog.String("error", err.Error()))
		// Return a simpler error type, as the underlying error from Execute might not be needed by caller
		return errors.New(errMsgForError)
	}

	presenter.Newline()
	presenter.Success("Terraform apply successful.")
	logger.Info("Terraform apply successful", slog.String("source_command", "deploy"))
	return nil
}

// executePulumiDeploy now accepts execClient
func executePulumiDeploy(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execDeployClientInterface, dir string, skipConfirm bool) error {
	tool := "pulumi"
	args := []string{"up"}

	if !execClient.CommandExists(tool) { // Use execClient
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Pulumi is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Pulumi deploy prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	// Confirmation logic remains the same
	presenter.Newline()
	presenter.Info("Proposed Deploy Action:")
	presenter.Detail("Run '%s %s'.", tool, strings.Join(args, " "))
	presenter.Detail("(Note: '%s up' will show its own preview and prompt for confirmation before making changes).", tool)
	presenter.Newline()

	confirmed := false
	if skipConfirm {
		presenter.Info("Confirmation prompt (for contextvibes) bypassed via --yes flag.")
		logger.InfoContext(ctx, "Wrapper confirmation bypassed via flag", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Bool("yes_flag", true))
		confirmed = true
	} else {
		var promptErr error
		confirmed, promptErr = presenter.PromptForConfirmation("Proceed to run 'pulumi up'?")
		if promptErr != nil {
			logger.ErrorContext(ctx, "Error reading deploy confirmation", slog.String("tool", tool), slog.String("error", promptErr.Error()))
			return promptErr
		}
	}

	if !confirmed {
		presenter.Info("'pulumi up' command aborted by user (before execution).")
		logger.InfoContext(ctx, "Deploy aborted by user confirmation", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Bool("confirmed", false))
		return nil
	}
	logger.DebugContext(ctx, "Proceeding after deploy confirmation", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Bool("confirmed", true))

	// Execution using execClient
	presenter.Newline()
	presenter.Info("Starting Pulumi execution ('%s %s')...", tool, strings.Join(args, " "))
	logger.Info("Executing pulumi up", slog.String("source_command", "deploy"), slog.String("tool", tool), slog.Any("args", args))

	// Use execClient.Execute - pulumi up pipes its own output
	err := execClient.Execute(ctx, dir, tool, args...) // Use execClient
	if err != nil {
		// Error message from Execute should contain exit code info.
		// User will see the piped output from pulumi up itself.
		errMsgForUser := "'pulumi up' command failed or was aborted by user during its execution."
		errMsgForError := "pulumi up command failed or aborted"
		presenter.Error(errMsgForUser)
		logger.Error("Pulumi up command failed or aborted", slog.String("source_command", "deploy"), slog.String("error", err.Error()))
		return errors.New(errMsgForError)
	}

	presenter.Newline()
	presenter.Success("Pulumi up completed successfully.")
	logger.Info("Pulumi up successful", slog.String("source_command", "deploy"))
	return nil
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

======== END FILE: cmd/deploy.go ========

======== FILE: cmd/describe.go ========

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

======== END FILE: cmd/describe.go ========

======== FILE: cmd/diff.go ========

// cmd/diff.go

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/git"   // Use GitClient
	"github.com/contextvibes/cli/internal/tools" // For Markdown/File IO
	"github.com/contextvibes/cli/internal/ui"    // Use Presenter for UI
	"github.com/spf13/cobra"
)

const fixedDiffOutputFile = "contextvibes.md" // Keep specific to diff command

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: fmt.Sprintf("Shows pending Git changes, OVERWRITING %s.", fixedDiffOutputFile),
	Long: fmt.Sprintf(`Generates a Markdown summary of pending changes (staged, unstaged, untracked)
in the Git repository and OVERWRITES the context file: %s.

This is useful for providing diff context to AI assistants or for quick status checks.
Run 'contextvibes describe' again if you need the full project context instead.`, fixedDiffOutputFile),
	Example: `  contextvibes diff  # OVERWRITES contextvibes.md with diff summary`,
	Args:    cobra.NoArgs,
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

		presenter.Summary("Generating Git diff summary for %s.", fixedDiffOutputFile)

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Diff: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "diff"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "diff"))

		// --- Generate Diff Content ---
		var outputBuffer bytes.Buffer // Buffer to build markdown content
		var hasChanges bool

		presenter.Step("Checking for staged changes...") // Use Step
		stagedOut, _, stagedErr := client.GetDiffCached(ctx)
		if stagedErr != nil {
			presenter.Error("Failed to get staged changes: %v", stagedErr)
			return stagedErr // Client logs details
		}
		stagedOut = strings.TrimSpace(stagedOut)
		if stagedOut != "" {
			hasChanges = true
			logger.DebugContext(ctx, "Adding staged changes to buffer", slog.String("source_command", "diff"))
			tools.AppendSectionHeader(&outputBuffer, "Staged Changes (Index / `git diff --cached`)")
			tools.AppendFencedCodeBlock(&outputBuffer, stagedOut, "diff")
		}

		presenter.Step("Checking for unstaged changes...") // Use Step
		unstagedOut, _, unstagedErr := client.GetDiffUnstaged(ctx)
		if unstagedErr != nil {
			presenter.Error("Failed to get unstaged changes: %v", unstagedErr)
			return unstagedErr // Client logs details
		}
		unstagedOut = strings.TrimSpace(unstagedOut)
		if unstagedOut != "" {
			hasChanges = true
			logger.DebugContext(ctx, "Adding unstaged changes to buffer", slog.String("source_command", "diff"))
			tools.AppendSectionHeader(&outputBuffer, "Unstaged Changes (Working Directory / `git diff HEAD`)")
			tools.AppendFencedCodeBlock(&outputBuffer, unstagedOut, "diff")
		}

		presenter.Step("Checking for untracked files...") // Use Step
		untrackedOut, _, untrackedErr := client.ListUntrackedFiles(ctx)
		if untrackedErr != nil {
			presenter.Error("Failed to list untracked files: %v", untrackedErr)
			return untrackedErr // Client logs details
		}
		untrackedOut = strings.TrimSpace(untrackedOut)
		if untrackedOut != "" {
			hasChanges = true
			logger.DebugContext(ctx, "Adding untracked files to buffer", slog.String("source_command", "diff"))
			tools.AppendSectionHeader(&outputBuffer, "Untracked Files (`git ls-files --others --exclude-standard`)")
			tools.AppendFencedCodeBlock(&outputBuffer, untrackedOut, "")
		}

		// --- Write Output File or Report No Changes ---
		presenter.Newline()
		if !hasChanges {
			presenter.Info("No pending changes found.")
			presenter.Advice("The context file '%s' remains unchanged.", fixedDiffOutputFile)
			logger.InfoContext(ctx, "No pending git changes detected.", slog.String("source_command", "diff"))
		} else {
			presenter.Step("Writing diff summary, overwriting %s...", fixedDiffOutputFile) // Use Step

			// tools.WriteBufferToFile currently prints its own messages.
			// If we want full control via Presenter, WriteBufferToFile would need
			// modification or replacement. Let's keep it for now.
			errWrite := tools.WriteBufferToFile(fixedDiffOutputFile, &outputBuffer)
			if errWrite != nil {
				presenter.Error("Failed to write output file '%s': %v", fixedDiffOutputFile, errWrite)
				logger.ErrorContext(ctx, "Failed to write diff output file" /*...*/)
				return errWrite
			}
			// Success message is printed by WriteBufferToFile currently.
			// If WriteBufferToFile is made silent:
			// presenter.Success("Successfully wrote diff summary to %s.", fixedDiffOutputFile)
			logger.InfoContext(ctx, "Successfully wrote git diff summary to file." /*...*/)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

======== END FILE: cmd/diff.go ========

======== FILE: cmd/format.go ========

// cmd/format.go
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/contextvibes/cli/internal/project"
	// "github.com/contextvibes/cli/internal/tools" // No longer needed for exec functions
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
	// No direct import of internal/exec needed if using global ExecClient from cmd/root.go
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Applies code formatting (go fmt, terraform fmt, isort, black).",
	Long: `Detects project type (Go, Python, Terraform) and applies standard formatting
using available tools in PATH, modifying files in place.

- Go: Runs 'go fmt ./...'
- Python: Runs 'isort .' and 'black .'.
- Terraform: Runs 'terraform fmt -recursive .'.

This command focuses only on applying formatting, unlike 'quality' which checks
formatters, linters, and validators.`,
	Example:       `  contextvibes format  # Apply formatting to Go, Python, or Terraform files`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		// Use global ExecClient from cmd/root.go
		if ExecClient == nil {
			return fmt.Errorf("internal error: executor client not initialized")
		}
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Applying code formatting.")

		cwd, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Format: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return err
		}

		presenter.Info("Detecting project type...")
		projType, err := project.Detect(cwd)
		if err != nil {
			logger.ErrorContext(ctx, "Format: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return err
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "format"), slog.String("type", string(projType)))

		hasTerraform := projType == project.Terraform
		hasPython := projType == project.Python
		hasGo := projType == project.Go

		if !hasTerraform && !hasPython && !hasGo {
			presenter.Info("No supported components (Terraform, Python, Go) found for formatting in this directory.")
			return nil
		}

		var formatErrors []error

		// --- Terraform Formatting ---
		if hasTerraform {
			presenter.Newline()
			presenter.Header("Terraform Formatting")
			tool := "terraform"
			if ExecClient.CommandExists(tool) { // Use ExecClient
				presenter.Step("Running terraform fmt...")
				logger.Info("Executing terraform fmt -recursive .", slog.String("source_command", "format"))
				// terraform fmt pipes its own output (files changed)
				errFmt := ExecClient.Execute(ctx, cwd, tool, "fmt", "-recursive", ".") // Use ExecClient
				if errFmt != nil {
					errMsg := fmt.Sprintf("`terraform fmt` failed or reported issues. Error: %v", errFmt)
					presenter.Error(errMsg)
					formatErrors = append(formatErrors, errors.New("terraform fmt failed"))
					logger.Error("Terraform fmt failed", slog.String("source_command", "format"), slog.String("error", errFmt.Error()))
				} else {
					presenter.Success("terraform fmt completed.")
					logger.Info("Terraform fmt successful", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Terraform formatting.", tool)
				presenter.Warning(msg)
				logger.Warn("Terraform format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", tool))
			}
		}

		// --- Python Formatting ---
		if hasPython {
			presenter.Newline()
			presenter.Header("Python Formatting")
			pythonDir := "."

			toolIsort := "isort"
			if ExecClient.CommandExists(toolIsort) { // Use ExecClient
				presenter.Step("Running %s...", toolIsort)
				logger.Info("Executing isort .", slog.String("source_command", "format"))
				errIsort := ExecClient.Execute(ctx, cwd, toolIsort, pythonDir) // Use ExecClient
				if errIsort != nil {
					errMsg := fmt.Sprintf("`%s` failed or reported issues. Error: %v", toolIsort, errIsort)
					presenter.Error(errMsg)
					formatErrors = append(formatErrors, errors.New("isort failed"))
					logger.Error("isort failed", slog.String("source_command", "format"), slog.String("error", errIsort.Error()))
				} else {
					presenter.Success("%s completed.", toolIsort)
					logger.Info("isort successful", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping import sorting.", toolIsort)
				presenter.Warning(msg)
				logger.Warn("isort format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolIsort))
			}

			toolBlack := "black"
			if ExecClient.CommandExists(toolBlack) { // Use ExecClient
				presenter.Step("Running %s...", toolBlack)
				logger.Info("Executing black .", slog.String("source_command", "format"))
				// Black exits 0 if no changes, 1 if reformatted, >1 on error.
				// ExecClient.Execute will return an error for non-zero exit.
				// We can interpret this: if no error, no changes. If error, could be reformat or actual fail.
				// For `format` command, successful reformatting is a success.
				// The `OSCommandExecutor` logs the exit code, so we can rely on its error message or check stderr.
				// For simplicity, we treat any non-zero exit from black as "files were formatted or error occurred".
				// The user sees black's direct output.
				errBlack := ExecClient.Execute(ctx, cwd, toolBlack, pythonDir) // Use ExecClient
				if errBlack != nil {
					// Check if it's just a reformatting (exit code 1 for black typically means files changed)
					// This requires more complex error inspection if we want to distinguish.
					// For now, if black exits non-zero, we log it as potentially having issues.
					// A more robust solution might use CaptureOutput and inspect exit code and stderr.
					errMsg := fmt.Sprintf("`%s` completed (may have reformatted files or encountered an issue). Error (if any): %v", toolBlack, errBlack)
					presenter.Info(errMsg) // Info, as reformatting is the goal. If actual error, black would show it.
					logger.Warn("black completed with non-zero exit", slog.String("source_command", "format"), slog.String("error", errBlack.Error()))
					// Don't add to formatErrors unless we are sure it's a critical failure, not just reformatting.
					// If it's a critical failure, black's output to stderr (piped by Execute) should indicate it.
				} else {
					presenter.Success("%s completed (no changes needed).", toolBlack)
					logger.Info("black successful (no changes)", slog.String("source_command", "format"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Python code formatting.", toolBlack)
				presenter.Warning(msg)
				logger.Warn("black format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolBlack))
			}
		}

		// --- Go Formatting ---
		if hasGo {
			presenter.Newline()
			presenter.Header("Go Formatting")
			goDir := "./..."

			toolGo := "go"
			if ExecClient.CommandExists(toolGo) { // Use ExecClient
				presenter.Step("Running go fmt...")
				logger.Info("Executing go fmt ./...", slog.String("source_command", "format"))
				// `go fmt` prints changed file paths to stdout.
				// We can use CaptureOutput to see if it did anything.
				stdout, stderr, errFmt := ExecClient.CaptureOutput(ctx, cwd, toolGo, "fmt", goDir) // Use ExecClient
				if errFmt != nil {
					errMsg := fmt.Sprintf("`go fmt` failed. Error: %v", errFmt)
					if stderr != "" {
						errMsg += fmt.Sprintf("\nStderr: %s", stderr)
					}
					presenter.Error(errMsg)
					formatErrors = append(formatErrors, errors.New("go fmt failed"))
					logger.Error("go fmt failed", slog.String("source_command", "format"), slog.String("error", errFmt.Error()), slog.String("stderr", stderr))
				} else {
					if stdout != "" {
						presenter.Success("go fmt completed and formatted the following files:")
						presenter.Detail(stdout) // Show which files were formatted
					} else {
						presenter.Success("go fmt completed (no files needed formatting).")
					}
					logger.Info("go fmt successful", slog.String("source_command", "format"), slog.String("stdout", stdout))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Go formatting.", toolGo)
				presenter.Warning(msg)
				logger.Warn("Go format skipped: command not found", slog.String("source_command", "format"), slog.String("tool", toolGo))
			}
		}

		presenter.Newline()
		presenter.Header("Formatting Summary")
		if len(formatErrors) > 0 {
			errMsg := fmt.Sprintf("%d formatting tool(s) reported errors.", len(formatErrors))
			presenter.Error(errMsg)
			presenter.Advice("Review the errors above.")
			logger.Error("Format command failed due to errors", slog.String("source_command", "format"), slog.Int("error_count", len(formatErrors)))
			return formatErrors[0]
		}

		presenter.Success("All formatting tools completed successfully or applied changes.")
		logger.Info("Format command finished", slog.String("source_command", "format"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(formatCmd)
}

======== END FILE: cmd/format.go ========

======== FILE: cmd/kickoff.go ========

// cmd/kickoff.go
package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/kickoff"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	branchNameFlag            string 
	isStrategicKickoffFlag  bool   
	markStrategicCompleteFlag bool   
)

var kickoffCmd = &cobra.Command{
	Use:   "kickoff [--branch <branch-name>] [--strategic] [--mark-strategic-complete]",
	Short: "Manages project kickoff: daily branch workflow or strategic project initiation.",
	Long: `Manages project kickoff workflows.

Default Behavior (Daily Kickoff, if strategic completed):
  - Requires a clean state on the main branch.
  - Updates the main branch, creates a new daily/feature branch, and pushes it.
  - Uses --branch flag or prompts for name (respects .contextvibes.yaml validation).

Strategic Kickoff Prompt Generation (--strategic, or if first run):
  - Initiates a brief interactive session to gather basic project details.
  - Generates a comprehensive master prompt file (STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md).
  - User takes this prompt to an external AI to complete the detailed strategic kickoff.

Marking Strategic Kickoff as Complete (--mark-strategic-complete):
  - Updates '.contextvibes.yaml' to indicate the strategic kickoff has been done.
  - This enables the daily kickoff workflow for subsequent runs without '--strategic'.

Global --yes flag (from root command) bypasses confirmations for daily kickoff actions.`,
	Example: `  # Daily Kickoff Examples (assumes strategic kickoff was previously marked complete)
  contextvibes kickoff --branch feature/new-login
  contextvibes kickoff -b fix/bug-123 -y
  contextvibes kickoff # Prompts for branch name

  # Strategic Kickoff Prompt Generation
  contextvibes kickoff --strategic 
  contextvibes kickoff             # Runs strategic prompt generation if first time

  # Mark Strategic Kickoff as Done (after user completes session with external AI)
  contextvibes kickoff --mark-strategic-complete`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		if LoadedAppConfig == nil {
			presenter.Error("Internal error: Application configuration not loaded.")
			logger.ErrorContext(ctx, "Kickoff cmd failed: LoadedAppConfig is nil", slog.String("source_command", "kickoff"))
			return errors.New("application configuration not loaded")
		}
		if ExecClient == nil {
			presenter.Error("Internal error: Executor client not initialized.")
			logger.ErrorContext(ctx, "Kickoff cmd failed: ExecClient is nil", slog.String("source_command", "kickoff"))
			return errors.New("executor client not initialized")
		}

		var configFilePath string
		// Attempt to find .contextvibes.yaml in repo root
		repoCfgPath, findPathErr := config.FindRepoRootConfigPath(ExecClient) 
		if findPathErr != nil { // Error finding repo root (e.g., not a git repo)
			logger.WarnContext(ctx, "Could not determine git repository root. '.contextvibes.yaml' will be looked for/created in current directory.",
				slog.String("source_command", "kickoff"), slog.Any("find_path_error", findPathErr))
			cwd, _ := os.Getwd()
			configFilePath = filepath.Join(cwd, config.DefaultConfigFileName)
		} else if repoCfgPath == "" { // Repo root found, but .contextvibes.yaml doesn't exist there
			logger.InfoContext(ctx, "'.contextvibes.yaml' not found in repository root. It will be created there if needed.",
				slog.String("source_command", "kickoff"))
			// Get repo root again to ensure correct path for creation
			repoRootForCreation, _, _ := ExecClient.CaptureOutput(context.Background(), ".", "git", "rev-parse", "--show-toplevel")
			cleanRoot := strings.TrimSpace(repoRootForCreation)
			if cleanRoot == "" || cleanRoot == "." { // Fallback if somehow rev-parse fails here
				cwd, _ := os.Getwd()
				cleanRoot = cwd
			}
			configFilePath = filepath.Join(cleanRoot, config.DefaultConfigFileName)
		} else { // Config file found in repo root
			configFilePath = repoCfgPath
		}
		logger.DebugContext(ctx, "Determined config file path for kickoff operations", 
			slog.String("path", configFilePath), slog.String("source_command", "kickoff"))


		workDir, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get working directory: %v", err)
			logger.ErrorContext(ctx, "Kickoff cmd: Failed getwd", slog.String("error", err.Error()))
			return err
		}
		
		var gitClt *git.GitClient
		gitClientConfig := git.GitClientConfig{
			Logger:                logger, 
			DefaultRemoteName:     LoadedAppConfig.Git.DefaultRemote,
			DefaultMainBranchName: LoadedAppConfig.Git.DefaultMainBranch,
			Executor:              ExecClient.UnderlyingExecutor(), 
		}
		// Try to initialize Git client. It's okay if it fails for some strategic operations.
		gitClt, err = git.NewClient(ctx, workDir, gitClientConfig)
		if err != nil {
			// Only log a warning here. The orchestrator will decide if a nil gitClient is fatal for the chosen operation.
			logger.WarnContext(ctx, "Kickoff cmd: Git client initialization failed. Some operations might be limited.", 
				slog.String("source_command", "kickoff"), 
				slog.String("error", err.Error()))
			// gitClt will be nil
		}

		// The global 'assumeYes' is set by rootCmd.PersistentPreRunE or by flag parsing.
		orchestrator := kickoff.NewOrchestrator(logger, LoadedAppConfig, presenter, gitClt, configFilePath)

		if markStrategicCompleteFlag {
			// User wants to mark the strategic kickoff as done.
			if isStrategicKickoffFlag || branchNameFlag != "" {
				presenter.Warning("--mark-strategic-complete is mutually exclusive with --strategic and --branch. Ignoring other flags.")
				logger.WarnContext(ctx, "Redundant flags with --mark-strategic-complete", 
					slog.Bool("strategic_flag", isStrategicKickoffFlag),
					slog.String("branch_flag", branchNameFlag))
			}
			err = orchestrator.MarkStrategicKickoffComplete(ctx)
		} else {
			// Pass the global 'assumeYes' (from cmd/root.go, bound to rootCmd.PersistentFlags())
			err = orchestrator.ExecuteKickoff(ctx, isStrategicKickoffFlag, branchNameFlag, assumeYes)
		}
		
		if err != nil {
			// Orchestrator's presenter methods should have already shown user-friendly messages.
			// Log the error from the orchestrator for the AI trace log.
			logger.ErrorContext(ctx, "Kickoff command execution resulted in error", 
				slog.String("source_command", "kickoff"), 
				slog.Any("error", err)) // err already includes context from orchestrator
			// Return error to signify failure to Cobra, avoids double-printing if presenter handled it.
			return err 
		}

		logger.InfoContext(ctx, "Kickoff command completed successfully.", slog.String("source_command", "kickoff"))
		return nil
	},
}

func init() {
	kickoffCmd.Flags().StringVarP(&branchNameFlag, "branch", "b", "", "Name for the new daily/feature branch (e.g., feature/JIRA-123)")
	kickoffCmd.Flags().BoolVar(&isStrategicKickoffFlag, "strategic", false, "Generates a master prompt for an AI-guided strategic project kickoff session.")
	kickoffCmd.Flags().BoolVar(&markStrategicCompleteFlag, "mark-strategic-complete", false, "Marks the strategic kickoff as complete in .contextvibes.yaml.")
	
	rootCmd.AddCommand(kickoffCmd)
}

======== END FILE: cmd/kickoff.go ========

======== FILE: cmd/plan.go ========

// cmd/plan.go
package cmd

import (
	"context" // Import context
	"errors"
	"fmt"
	"log/slog" // For logging errors to AI file
	"os"
	osexec "os/exec" // Alias for standard library exec.ExitError, in case of ExitError check
	"strings"

	"github.com/contextvibes/cli/internal/project"
	// "github.com/contextvibes/cli/internal/tools" // No longer needed for exec functions
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
	// Ensure internal/exec is available if not already imported by other files in cmd
	// but we'll be using the global ExecClient from cmd/root.go
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generates an execution plan (e.g., terraform plan, pulumi preview).",
	Long: `Detects the project type (Terraform, Pulumi) and runs the appropriate
command to generate an execution plan, showing expected infrastructure changes.

- Terraform: Runs 'terraform plan -out=tfplan.out'
- Pulumi: Runs 'pulumi preview'`,
	Example:       `  contextvibes plan  # Run in a Terraform or Pulumi project directory`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger // From cmd/root.go
		// ExecClient should also be available from cmd/root.go
		if ExecClient == nil {
			return fmt.Errorf("internal error: executor client not initialized")
		}
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}

		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Generating execution plan.")

		cwd, err := os.Getwd()
		if err != nil {
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Plan: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return wrappedErr
		}

		presenter.Info("Detecting project type in %s...", presenter.Highlight(cwd))
		projType, err := project.Detect(cwd)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to detect project type: %w", err)
			logger.ErrorContext(ctx, "Plan: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return wrappedErr
		}

		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "plan"), slog.String("type", string(projType)))

		switch projType {
		case project.Terraform:
			return executeTerraformPlan(ctx, presenter, logger, ExecClient, cwd) // Pass ExecClient
		case project.Pulumi:
			return executePulumiPreview(ctx, presenter, logger, ExecClient, cwd) // Pass ExecClient
		case project.Go:
			presenter.Info("Plan command is not applicable for Go projects.")
			return nil
		case project.Python:
			presenter.Info("Plan command is not applicable for Python projects.")
			return nil
		case project.Unknown:
			errMsgForUser := "Unknown project type detected. Cannot determine plan action."
			errMsgForError := "unknown project type detected"
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "plan"))
			return errors.New(errMsgForError)
		default:
			errMsgForUser := fmt.Sprintf("Internal error: Unhandled project type '%s'", projType)
			errMsgForError := fmt.Sprintf("internal error: unhandled project type '%s'", projType)
			presenter.Error(errMsgForUser)
			logger.Error(errMsgForUser, slog.String("source_command", "plan"))
			return errors.New(errMsgForError)
		}
	},
}

// Modified to accept execClient
func executeTerraformPlan(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execClientInterface, dir string) error {
	tool := "terraform"
	args := []string{"plan", "-out=tfplan.out"}

	if !execClient.CommandExists(tool) { // Use ExecClient
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Terraform is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Terraform plan prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	presenter.Info("Executing: %s %s", tool, strings.Join(args, " "))
	logger.Info("Executing terraform plan", slog.String("source_command", "plan"), slog.String("tool", tool), slog.Any("args", args))

	// Use ExecClient.Execute. Since terraform plan's output is important, CaptureOutput might be better,
	// but if the OSCommandExecutor pipes stdio for Execute, it might be fine.
	// For consistency with how it might have worked before (piping output), Execute is okay.
	// However, to check exit codes correctly, CaptureOutput is often more robust as Execute's error might be too generic.
	// Let's switch to CaptureOutput to analyze exit codes more reliably.
	_, stderr, err := execClient.CaptureOutput(ctx, dir, tool, args...) // Use CaptureOutput

	if err != nil {
		var exitErr *osexec.ExitError // from os/exec
		if errors.As(err, &exitErr) {
			// Exit code 2 from `terraform plan` means changes are needed (success for plan command)
			if exitErr.ExitCode() == 2 {
				presenter.Newline()
				// stderr might contain the plan output itself or useful info, so display it.
				if strings.TrimSpace(stderr) != "" {
					presenter.Detail("Terraform plan output (stderr):\n%s", stderr)
				}
				presenter.Info("Terraform plan indicates changes are needed.")
				presenter.Advice("Plan saved to tfplan.out. Run `contextvibes deploy` to apply.")
				logger.Info("Terraform plan successful (changes detected)", slog.String("source_command", "plan"), slog.Int("exit_code", 2))
				return nil
			}
			// Any other non-zero exit code is a failure
			errMsgForUser := fmt.Sprintf("'%s plan' command failed.", tool)
			errMsgForError := fmt.Sprintf("%s plan command failed", tool)
			presenter.Error(errMsgForUser)
			if strings.TrimSpace(stderr) != "" {
				presenter.Error("Details (stderr):\n%s", stderr)
			}
			logger.Error("Terraform plan command failed", slog.String("source_command", "plan"), slog.Int("exit_code", exitErr.ExitCode()), slog.String("error", err.Error()), slog.String("stderr", stderr))
			return errors.New(errMsgForError)
		}
		// Error wasn't an ExitError
		errMsgForUser := fmt.Sprintf("Failed to execute '%s plan': %v", tool, err)
		presenter.Error(errMsgForUser)
		if strings.TrimSpace(stderr) != "" {
			presenter.Error("Details (stderr):\n%s", stderr)
		}
		logger.Error("Terraform plan execution failed", slog.String("source_command", "plan"), slog.String("error", err.Error()), slog.String("stderr", stderr))
		return fmt.Errorf("failed to execute '%s plan': %w", tool, err)
	}

	// Exit code 0 means no changes detected
	// stdout from `terraform plan -out=...` is usually minimal, confirmation messages.
	// The actual plan is in the file or on stderr if not using -out.
	presenter.Newline()
	presenter.Info("Terraform plan successful (no changes detected).")
	presenter.Advice("Plan saved to tfplan.out (contains no changes).")
	logger.Info("Terraform plan successful (no changes)", slog.String("source_command", "plan"), slog.Int("exit_code", 0))
	return nil
}

// Modified to accept execClient
func executePulumiPreview(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execClientInterface, dir string) error {
	tool := "pulumi"
	args := []string{"preview"}

	if !execClient.CommandExists(tool) { // Use ExecClient
		errMsgForUser := fmt.Sprintf("Command '%s' not found. Please ensure Pulumi is installed and in your PATH.", tool)
		errMsgForError := fmt.Sprintf("command '%s' not found", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Pulumi preview prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New(errMsgForError)
	}

	presenter.Info("Executing: %s %s", tool, strings.Join(args, " "))
	logger.Info("Executing pulumi preview", slog.String("source_command", "plan"), slog.String("tool", tool), slog.Any("args", args))

	// Pulumi preview prints to stdout/stderr itself.
	// Using ExecClient.Execute will pipe these streams.
	err := execClient.Execute(ctx, dir, tool, args...)
	if err != nil {
		// Error message from ExecClient.Execute should be informative enough
		// (includes exit code if that's the cause)
		errMsgForUser := fmt.Sprintf("'%s preview' command failed.", tool)
		errMsgForError := fmt.Sprintf("%s preview command failed", tool)
		presenter.Error(errMsgForUser) // The actual error details would have been piped to stderr by Pulumi
		logger.Error("Pulumi preview command failed", slog.String("source_command", "plan"), slog.String("error", err.Error()))
		return errors.New(errMsgForError)
	}

	presenter.Newline()
	presenter.Success("Pulumi preview completed successfully.")
	logger.Info("Pulumi preview successful", slog.String("source_command", "plan"))
	return nil
}

// Define an interface for execClient to make testing/mocking easier for these functions.
// This interface matches the methods used from exec.ExecutorClient.
type execClientInterface interface {
	CommandExists(commandName string) bool
	Execute(ctx context.Context, dir string, commandName string, args ...string) error
	CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (string, string, error)
}

func init() {
	rootCmd.AddCommand(planCmd)
}

======== END FILE: cmd/plan.go ========

======== FILE: cmd/quality.go ========

// cmd/quality.go

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings" // Added for trimming space

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui" // Use Presenter
	"github.com/spf13/cobra"
)

// Assume AppLogger is initialized in rootCmd for AI file logging
// var AppLogger *slog.Logger // Defined in root.go

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "Runs code formatting and linting checks (Terraform, Python, Go).",
	Long: `Detects project type (Terraform, Python, Go) and runs common formatters and linters.
Checks performed depend on available tools in PATH.

- Terraform: Runs 'terraform fmt -check', 'terraform validate', 'tflint'.
- Python: Runs 'isort --check', 'black --check', 'flake8'.
- Go: Runs 'go vet', 'go mod tidy', and checks 'go fmt' compliance.

Formatter/validator checks ('terraform fmt -check', 'terraform validate', 'isort --check',
'black --check', 'go vet') and dependency checks ('go mod tidy') will fail the command
if issues are found or errors occur.
The 'go fmt' check will also fail the command if files *are not* correctly formatted
(note: this check modifies files in place if needed to determine compliance).
Linter issues ('tflint', 'flake8') are reported as warnings.`,
	Example: `  contextvibes quality`,
	Args:    cobra.NoArgs,
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

		presenter.Summary("Running code quality checks.")

		cwd, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Quality: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return err
		}

		presenter.Info("Detecting project type...")
		projType, err := project.Detect(cwd)
		if err != nil {
			logger.ErrorContext(ctx, "Quality: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return err
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "quality"), slog.String("type", string(projType)))

		hasTerraform := projType == project.Terraform
		hasPython := projType == project.Python
		hasGo := projType == project.Go

		if !hasTerraform && !hasPython && !hasGo {
			presenter.Info("No supported components (Terraform, Python, Go) found for quality checks in this directory.")
			return nil
		}

		var criticalErrors []error
		var warnings []string

		// TODO: Offload specific checks for each project type into dedicated internal functions/packages.

		// --- Terraform Checks ---
		if hasTerraform {
			presenter.Newline()
			presenter.Header("Terraform Quality Checks")
			tool := "terraform"
			if ExecClient.CommandExists(tool) {
				// --- terraform fmt -check ---
				presenter.Step("Checking Terraform formatting (terraform fmt -check)...")
				logger.Info("Executing terraform fmt -check -recursive .", slog.String("source_command", "quality"))
				errFmt := ExecClient.Execute(ctx, cwd, tool, "fmt", "-check", "-recursive", ".")
				if errFmt != nil {
					errMsg := "`terraform fmt -check` failed or found files needing formatting"
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errFmt))
					criticalErrors = append(criticalErrors, errors.New("terraform fmt check failed"))
					logger.Error("Terraform fmt check failed", slog.String("source_command", "quality"), slog.String("error", errFmt.Error()))
					presenter.Advice("Run `contextvibes format` or `terraform fmt -recursive .` to fix formatting.")
				} else {
					presenter.Success("terraform fmt check completed (no changes needed).")
					logger.Info("Terraform fmt check successful", slog.String("source_command", "quality"))
				}

				// --- terraform validate ---
				if errFmt == nil { // Skip if formatting failed
					presenter.Step("Running terraform validate...")
					logger.Info("Executing terraform validate", slog.String("source_command", "quality"))
					errValidate := ExecClient.Execute(ctx, cwd, tool, "validate")
					if errValidate != nil {
						errMsg := "`terraform validate` failed"
						presenter.Error(errMsg)
						warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errValidate))
						criticalErrors = append(criticalErrors, errors.New("terraform validate failed"))
						logger.Error("Terraform validate failed", slog.String("source_command", "quality"), slog.String("error", errValidate.Error()))
					} else {
						presenter.Success("terraform validate completed.")
						logger.Info("Terraform validate successful", slog.String("source_command", "quality"))
					}
				} else {
					presenter.Warning("Skipping terraform validate due to previous terraform fmt check failure.")
					logger.Warn("Skipping terraform validate due to fmt failure", slog.String("source_command", "quality"))
				}
			} else {
				// Handle missing terraform tool
				msg := fmt.Sprintf("'%s' command not found, skipping Terraform format/validate.", tool)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("Terraform checks skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", tool))
			}

			// --- tflint ---
			linter := "tflint"
			if ExecClient.CommandExists(linter) {
				presenter.Step("Running %s...", linter)
				logger.Info("Executing tflint", slog.String("source_command", "quality"))
				errLint := ExecClient.Execute(ctx, cwd, linter, "--recursive", ".")
				if errLint != nil {
					errMsg := fmt.Sprintf("`%s` reported issues or failed", linter)
					presenter.Warning(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errLint)) // Add to warnings, not critical
					logger.Warn("tflint reported issues or failed", slog.String("source_command", "quality"), slog.String("error", errLint.Error()))
				} else {
					presenter.Success("%s completed (no issues found).", linter)
					logger.Info("tflint successful", slog.String("source_command", "quality"))
				}
			} else {
				presenter.Info("'%s' command not found, skipping Terraform linting.", linter) // Info, as it's just a linter
				logger.Info("tflint check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", linter))
			}
		}

		// --- Python Checks ---
		if hasPython {
			presenter.Newline()
			presenter.Header("Python Quality Checks")
			pythonDir := "." // Assuming checks run from root

			// --- isort --check ---
			toolIsort := "isort"
			if ExecClient.CommandExists(toolIsort) {
				presenter.Step("Checking import sorting (%s --check)...", toolIsort)
				logger.Info("Executing isort --check", slog.String("source_command", "quality"))
				errIsort := ExecClient.Execute(ctx, cwd, toolIsort, "--check", pythonDir)
				if errIsort != nil {
					errMsg := fmt.Sprintf("`%s --check` failed or found files needing sorting", toolIsort)
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errIsort))
					criticalErrors = append(criticalErrors, errors.New("isort check failed"))
					logger.Error("isort check failed", slog.String("source_command", "quality"), slog.String("error", errIsort.Error()))
					presenter.Advice("Run `contextvibes format` or `isort .` to fix import sorting.")
				} else {
					presenter.Success("%s check completed (imports sorted).", toolIsort)
					logger.Info("isort check successful", slog.String("source_command", "quality"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping import sorting check.", toolIsort)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("isort check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", toolIsort))
			}

			// --- black --check ---
			toolBlack := "black"
			if ExecClient.CommandExists(toolBlack) {
				isortCheckFailed := false
				for _, e := range criticalErrors {
					if e.Error() == "isort check failed" {
						isortCheckFailed = true
						break
					}
				}
				if !isortCheckFailed {
					presenter.Step("Checking Python formatting (%s --check)...", toolBlack)
					logger.Info("Executing black --check", slog.String("source_command", "quality"))
					errBlack := ExecClient.Execute(ctx, cwd, toolBlack, "--check", pythonDir)
					if errBlack != nil {
						errMsg := fmt.Sprintf("`%s --check` failed or found files needing formatting", toolBlack)
						presenter.Error(errMsg)
						warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errBlack))
						criticalErrors = append(criticalErrors, errors.New("black check failed"))
						logger.Error("black check failed", slog.String("source_command", "quality"), slog.String("error", errBlack.Error()))
						presenter.Advice("Run `contextvibes format` or `black .` to fix formatting.")
					} else {
						presenter.Success("%s check completed (no changes needed).", toolBlack)
						logger.Info("black check successful", slog.String("source_command", "quality"))
					}
				} else {
					presenter.Warning("Skipping %s check due to previous python tool failure.", toolBlack)
					logger.Warn("Skipping black check due to prior failure", slog.String("source_command", "quality"))
				}
			} else {
				msg := fmt.Sprintf("'%s' command not found, skipping Python formatting check.", toolBlack)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("black check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", toolBlack))
			}

			// --- flake8 ---
			linterFlake8 := "flake8"
			if ExecClient.CommandExists(linterFlake8) {
				presenter.Step("Running %s...", linterFlake8)
				logger.Info("Executing flake8", slog.String("source_command", "quality"))
				errFlake8 := ExecClient.Execute(ctx, cwd, linterFlake8, pythonDir)
				if errFlake8 != nil {
					errMsg := fmt.Sprintf("`%s` reported issues or failed", linterFlake8)
					presenter.Warning(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errFlake8)) // Warning only
					logger.Warn("flake8 reported issues or failed", slog.String("source_command", "quality"), slog.String("error", errFlake8.Error()))
				} else {
					presenter.Success("%s completed (no issues found).", linterFlake8)
					logger.Info("flake8 successful", slog.String("source_command", "quality"))
				}
			} else {
				presenter.Info("'%s' command not found, skipping Python linting.", linterFlake8)
				logger.Info("flake8 check skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", linterFlake8))
			}
		}

		// --- Go Checks ---
		if hasGo {
			presenter.Newline()
			presenter.Header("Go Quality Checks")
			goDir := "./..." // Target all subdirectories for Go tools

			toolGo := "go"
			if ExecClient.CommandExists(toolGo) {

				// --- go fmt ---
				// Check formatting compliance by running `go fmt` and capturing output.
				// Note: This command *modifies files in place* if they are not formatted.
				// It acts as both a check and a fix within the quality command for Go fmt.
				// We treat non-empty output (files were formatted) as a critical error
				// because the code was not compliant before the command ran.
				// TODO: Revisit this if a reliable check-only mode becomes standard in go fmt
				//       or if external formatters (like gofumpt -l) are adopted.
				presenter.Step("Checking Go formatting (running go fmt)...")
				logger.Info("Executing go fmt ./... (and checking output)", slog.String("source_command", "quality"))
				fmtOutput, fmtStderr, errFmt := ExecClient.CaptureOutput(ctx, cwd, toolGo, "fmt", goDir)

				// First, check for execution errors (e.g., syntax errors).
				if errFmt != nil {
					errMsg := "`go fmt` execution failed"
					presenter.Error(errMsg + ": " + errFmt.Error())
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errFmt))             // Log warning for summary
					criticalErrors = append(criticalErrors, errors.New("go fmt execution failed")) // Critical error
					logger.Error("go fmt execution failed", slog.String("source_command", "quality"), slog.String("error", errFmt.Error()), slog.String("stderr", fmtStderr))
				} else {
					// If execution succeeded, check if files were actually formatted (non-empty output).
					trimmedOutput := strings.TrimSpace(fmtOutput)
					if trimmedOutput != "" {
						// Files *were* formatted, meaning they were not compliant initially. Treat as critical error.
						errMsg := "Go files were not correctly formatted (fixed by `go fmt`)"
						presenter.Error(errMsg)
						warnings = append(warnings, errMsg)                                             // Add to warnings for summary visibility
						criticalErrors = append(criticalErrors, errors.New("go fmt compliance failed")) // Critical error
						logger.Error("go fmt compliance failed: files were modified", slog.String("source_command", "quality"), slog.String("files_formatted", trimmedOutput))
						presenter.Advice("Commit the formatting changes applied by `go fmt`.")
						// Optional: show which files were formatted using presenter.Detail(trimmedOutput)
					} else {
						// Success: command ran without error and produced no output, meaning files were already formatted.
						presenter.Success("go fmt check completed (files already formatted).")
						logger.Info("go fmt check successful (no changes needed)", slog.String("source_command", "quality"))
					}
				}

				// --- go vet ---
				// Checks for suspicious constructs. Failure is critical.
				presenter.Step("Running Go vet...")
				logger.Info("Executing go vet ./...", slog.String("source_command", "quality"))
				errVet := ExecClient.Execute(ctx, cwd, toolGo, "vet", goDir)
				if errVet != nil {
					errMsg := "`go vet` reported issues or failed"
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errVet))
					criticalErrors = append(criticalErrors, errors.New("go vet failed"))
					logger.Error("go vet reported issues or failed", slog.String("source_command", "quality"), slog.String("error", errVet.Error()))
				} else {
					presenter.Success("go vet completed (no issues found).")
					logger.Info("go vet successful", slog.String("source_command", "quality"))
				}

				// --- go mod tidy ---
				// Ensures go.mod and go.sum are consistent. Failure is critical.
				// TODO: Add check if go.mod or go.sum were modified by tidy, potentially make that a critical error too.
				//       This likely requires checking git status before/after or diffing files.
				presenter.Step("Running go mod tidy...")
				logger.Info("Executing go mod tidy", slog.String("source_command", "quality"))
				errTidy := ExecClient.Execute(ctx, cwd, toolGo, "mod", "tidy")
				if errTidy != nil {
					errMsg := "`go mod tidy` failed"
					presenter.Error(errMsg)
					warnings = append(warnings, fmt.Sprintf("%s: %v", errMsg, errTidy))
					criticalErrors = append(criticalErrors, errors.New("go mod tidy failed"))
					logger.Error("go mod tidy failed", slog.String("source_command", "quality"), slog.String("error", errTidy.Error()))
				} else {
					presenter.Success("go mod tidy completed.")
					logger.Info("go mod tidy successful", slog.String("source_command", "quality"))
				}

			} else {
				// Handle missing 'go' tool
				msg := fmt.Sprintf("'%s' command not found, skipping Go quality checks.", toolGo)
				presenter.Warning(msg)
				warnings = append(warnings, msg)
				logger.Warn("Go checks skipped: command not found", slog.String("source_command", "quality"), slog.String("tool", toolGo))
			}
		}

		// --- Summary ---
		presenter.Newline()
		presenter.Header("Quality Checks Summary")

		if len(warnings) > 0 {
			presenter.Warning("Issues reported during checks (Includes non-critical linter findings and Go fmt results):") // Clarify scope
			for _, w := range warnings {
				presenter.Warning("  - %s", w)
			}
			presenter.Newline()
		}

		if len(criticalErrors) > 0 {
			errMsg := fmt.Sprintf("%d critical quality check(s) failed.", len(criticalErrors))
			presenter.Error(errMsg)
			presenter.Advice("Please review the errors above and fix them.")
			logger.Error("Quality command failed due to critical errors", slog.String("source_command", "quality"), slog.Int("error_count", len(criticalErrors)))
			return criticalErrors[0] // Return first critical error
		}

		// If there were warnings but no critical errors
		if len(warnings) > 0 {
			presenter.Success("All critical quality checks passed, but warnings were reported (check summary).")
		} else {
			presenter.Success("All quality checks passed successfully.")
		}
		logger.Info("Quality command finished", slog.String("source_command", "quality"), slog.Int("critical_errors", len(criticalErrors)), slog.Int("warnings_count", len(warnings)))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(qualityCmd)
}

======== END FILE: cmd/quality.go ========

======== FILE: cmd/root.go ========

package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/spf13/cobra"
)

// These are package-level variables, intended to be accessible by all files
// within this 'cmd' package.
var (
	AppLogger       *slog.Logger
	LoadedAppConfig *config.Config
	ExecClient      *exec.ExecutorClient // For general command execution
	assumeYes       bool                 // For --yes flag
	AppVersion      string               // Set in init() or by ldflags
)

// rootCmd is the base for all commands. It's also a package-level variable.
var rootCmd = &cobra.Command{
	Use:   "contextvibes",
	Short: "Manages project tasks: AI context generation, Git workflow, IaC, etc.",
	Long: `ContextVibes: Your Project Development Assistant CLI.

Designed to enhance your development workflow, ContextVibes offers a suite of
commands that bring consistency, automation, and AI-readiness to your daily tasks.

Key Capabilities:
  * Git Workflow Automation: Streamlined commands like 'kickoff', 'commit',
    'sync', 'wrapup', and 'status'. Features configurable validation for
    branch names and commit messages.
  * AI Context Generation: The 'describe' and 'diff' commands produce
    AI-friendly markdown ('contextvibes.md') detailing project state or
    changes, perfect for integrating with large language models.
  * Infrastructure & Code Management: Consistent wrappers for 'plan', 'deploy',
    'init' (IaC for Terraform/Pulumi), 'quality' checks, 'format' (code
    formatting for Go, Python, Terraform), and 'test' (project testing).
  * Programmatic Refactoring: The 'codemod' command allows applying
    structured code modifications from a JSON script.

Output & Logging for Clarity and AI:
  * User-Focused Terminal Output: Employs clear, structured messages with
    semantic prefixes (SUMMARY, INFO, ERROR, ADVICE, +, ~, !) and colors,
    all managed by an internal UI presenter.
  * Detailed AI Trace Log: Generates a separate, comprehensive JSON log
    (default: 'contextvibes_ai_trace.log', configurable) capturing in-depth
    execution details, ideal for AI analysis or advanced debugging.

Global Features for Control & Customization:
  * Non-Interactive Mode: Use the global '--yes' (or '-y') flag to
    automatically confirm prompts, enabling use in scripts and automation.
  * Project-Specific Configuration: Tailor default behaviors such as Git
    remote/main branch names, validation rule patterns (for branches and
    commits), and the default AI log file path using a '.contextvibes.yaml'
    file in your project's root directory.

For detailed information on any command, use 'contextvibes [command] --help'.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Temporary logger for bootstrap phase, before full config is loaded.
		// This logger writes to stderr for messages during the configuration loading process.
		tempLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

		// Minimal executor for finding config file, uses a discard logger
		// to avoid polluting logs before the main logger is set up.
		bootstrapOSExecutor := exec.NewOSCommandExecutor(slog.New(slog.NewTextHandler(io.Discard, nil)))
		bootstrapExecClient := exec.NewClient(bootstrapOSExecutor)

		defaultCfg := config.GetDefaultConfig()
		var loadedUserConfig *config.Config
		var configLoadErr error
		var foundConfigPath string

		// Attempt to find and load config file. The config is loaded from the repository root
		// or defaults to the CLI's internal configuration if none is found or there is an error during loading.
		repoConfigPath, findPathErr := config.FindRepoRootConfigPath(bootstrapExecClient)
		if findPathErr != nil {
			// Log to tempLogger (stderr) if finding path fails. The application will continue with default settings.
			tempLogger.Debug("Could not find git repo root to look for .contextvibes.yaml, using defaults.", slog.String("error", findPathErr.Error()))
			LoadedAppConfig = defaultCfg
		} else if repoConfigPath == "" {
			tempLogger.Debug(".contextvibes.yaml not found in repository root, using default configuration.")
			LoadedAppConfig = defaultCfg
		} else {
			foundConfigPath = repoConfigPath
			tempLogger.Debug("Attempting to load config file", slog.String("path", foundConfigPath))
			loadedUserConfig, configLoadErr = config.LoadConfig(foundConfigPath)

			if configLoadErr != nil {
				// User-facing warning to stderr if config loading fails
				fmt.Fprintf(os.Stderr, "[WARNING] Error loading config file '%s': %v. Using default settings.\n", foundConfigPath, configLoadErr)
				// Log to tempLogger (stderr) for more detail
				tempLogger.Error("Failed to load or parse .contextvibes.yaml, using defaults.", slog.String("path", foundConfigPath), slog.String("error", configLoadErr.Error()))
				LoadedAppConfig = defaultCfg
			} else if loadedUserConfig == nil {
				// Config file path was found, but file was empty or didn't parse to anything
				tempLogger.Info(".contextvibes.yaml was checked but not found or effectively empty, using default configuration.", slog.String("path_checked", foundConfigPath))
				LoadedAppConfig = defaultCfg
			} else {
				// Successfully loaded user config, now merge with defaults
				tempLogger.Info("Successfully loaded .contextvibes.yaml.", slog.String("path", foundConfigPath))
				LoadedAppConfig = config.MergeWithDefaults(loadedUserConfig, defaultCfg)
			}
		}

		// Determine AI log level and file path.
		// 1. Command-line flags take precedence over everything.
		// 2. If no flag is provided, the configuration file settings take precedence over the defaults.
		// 3. If neither a flag nor a config file setting is present, the built-in default values are used.
		aiLevel := parseLogLevel(logLevelAIValue, slog.LevelDebug) // logLevelAIValue is from the flag

		targetAILogFile := LoadedAppConfig.Logging.DefaultAILogFile // From merged config (or default if no user config)
		if aiLogFileFlagValue != "" {                               // aiLogFileFlagValue is from the flag
			targetAILogFile = aiLogFileFlagValue // Flag overrides config
		}

		// Initialize AppLogger (the main AI trace logger).
		var aiOut io.Writer = io.Discard // Default to discard if file opening fails
		logFileHandle, errLogFile := os.OpenFile(targetAILogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if errLogFile != nil {
			// User-facing error if AI log file cannot be opened
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to open AI log file '%s': %v. AI logs will be discarded.\n", targetAILogFile, errLogFile)
		} else {
			aiOut = logFileHandle
			// defer logFileHandle.Close() // This would close too early, needs to be closed on app exit if managed here. Usually handled by OS.
		}
		aiHandlerOptions := &slog.HandlerOptions{Level: aiLevel}
		aiHandler := slog.NewJSONHandler(aiOut, aiHandlerOptions)
		AppLogger = slog.New(aiHandler)

		// Initialize the main ExecutorClient with the now-configured AppLogger.
		mainOSExecutor := exec.NewOSCommandExecutor(AppLogger)
		ExecClient = exec.NewClient(mainOSExecutor)

		// Log initial setup to the now active AppLogger.
		AppLogger.Debug("AI Logger and main ExecutorClient initialized",
			slog.String("log_level_set_for_ai_file", aiLevel.String()),
			slog.String("ai_log_file_target", targetAILogFile),
			slog.Bool("ai_log_file_active", aiOut != io.Discard),
		)
		if assumeYes {
			AppLogger.Info("Running in non-interactive mode (--yes specified)")
		}

		if LoadedAppConfig != nil {
			// Log the effective configuration that the application will use.
			// Correctly determine the effective boolean value for validation settings:
			// Enabled if nil (not set by user, use default=true) OR if set to true by user.
			branchNameValidationEnabled := LoadedAppConfig.Validation.BranchName.Enable == nil || *LoadedAppConfig.Validation.BranchName.Enable
			commitMsgValidationEnabled := LoadedAppConfig.Validation.CommitMessage.Enable == nil || *LoadedAppConfig.Validation.CommitMessage.Enable

			AppLogger.Debug("Effective application configuration resolved",
				slog.Group("config",
					slog.Group("git",
						slog.String("defaultRemote", LoadedAppConfig.Git.DefaultRemote),
						slog.String("defaultMainBranch", LoadedAppConfig.Git.DefaultMainBranch),
					),
					slog.Group("logging",
						slog.String("defaultAILogFile", LoadedAppConfig.Logging.DefaultAILogFile),
					),
					slog.Group("validation",
						slog.Group("branchName",
							slog.Bool("enable", branchNameValidationEnabled),
							slog.String("pattern", LoadedAppConfig.Validation.BranchName.Pattern),
						),
						slog.Group("commitMessage",
							slog.Bool("enable", commitMsgValidationEnabled),
							slog.String("pattern", LoadedAppConfig.Validation.CommitMessage.Pattern),
						),
					),
				),
			)
		} else {
			// This should ideally not happen if logic above is correct
			AppLogger.Error("CRITICAL: LoadedAppConfig is unexpectedly nil after initialization attempt.")
			// Potentially return an error here to prevent CLI from running with no config
			// return errors.New("critical error: application configuration failed to load")
		}
		return nil
	},
}

// Execute is the main entry point for the CLI. It's made public so main.go can call it.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Ensure AppLogger is checked for nil before use, in case Execute() fails very early
		if AppLogger != nil {
			AppLogger.Error("CLI execution finished with error", slog.String("error", err.Error()))
		} else {
			// Fallback to stderr if logger isn't initialized
			fmt.Fprintf(os.Stderr, "[ERROR] CLI execution failed before logger initialization: %v\n", err)
		}
		os.Exit(1)
	}
}

// Flag variables should have distinct names from package-level vars if they are only for binding.
var (
	logLevelAIValue    string // Bound to --log-level-ai flag
	aiLogFileFlagValue string // Bound to --ai-log-file flag
)

func init() {
	// Set the application version. This can be overridden by ldflags during build.
	if AppVersion == "" {
		AppVersion = "v0.0.6" // Default version if not set by build flags
	}

	// Define persistent flags available to all commands.
	// Use different names for flag-bound variables (logLevelAIValue, aiLogFileFlagValue)
	// to avoid confusion with package-level variables that might be intended for direct use or derived values.
	// These flags bind to the variables logLevelAIValue and aiLogFileFlagValue, and take precedence over config file settings.
	rootCmd.PersistentFlags().StringVar(&logLevelAIValue, "log-level-ai", "debug", "AI (JSON) file log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&aiLogFileFlagValue, "ai-log-file", "",
		fmt.Sprintf("AI (JSON) log file path (overrides config default: see .contextvibes.yaml, fallback: %s)", config.UltimateDefaultAILogFilename))
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "Assume 'yes' to all confirmation prompts, enabling non-interactive mode")

	// Subcommands (like versionCmd, kickoffCmd, codemodCmd, etc.) add themselves to rootCmd
	// via their own init() functions. This is a standard Cobra pattern and keeps this file cleaner.
}

// parseLogLevel converts a string log level to an slog.Level.
func parseLogLevel(levelStr string, defaultLevel slog.Level) slog.Level {
	levelStrLower := strings.ToLower(strings.TrimSpace(levelStr))
	switch levelStrLower {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		// If an invalid level string is provided (and it's not empty/default),
		// print a warning to stderr.
		if levelStr != "" && !strings.EqualFold(levelStr, defaultLevel.String()) {
			fmt.Fprintf(os.Stderr, "[WARNING] Invalid AI log level '%s' provided. Using default level '%s'.\n", levelStr, defaultLevel.String())
		}
		return defaultLevel
	}
}

======== END FILE: cmd/root.go ========

======== FILE: cmd/status.go ========

// cmd/status.go

package cmd

import (
	"bufio" // For scanning output lines
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings" // For trimming space

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui" // Import presenter
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Shows a concise summary of the working tree status.",
	Long: `Displays a concise summary of the Git working tree status using 'git status --short'.
This includes staged changes, unstaged changes, and untracked files.`,
	Example: `  contextvibes status`,
	Args:    cobra.NoArgs,
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

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Status: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "status"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "status"))

		// --- Get and Display Short Status ---
		presenter.Summary("Displaying Git repository status summary.") // User info

		logger.DebugContext(ctx, "Fetching short status", slog.String("source_command", "status"))
		stdout, stderr, err := client.GetStatusShort(ctx)

		// Log stderr from git command if any (usually empty for status --short unless error)
		if stderr != "" {
			logger.WarnContext(ctx, "stderr received from 'git status --short'",
				slog.String("source_command", "status"),
				slog.String("stderr", strings.TrimSpace(stderr)),
			)
		}

		// Handle execution errors
		if err != nil {
			presenter.Error("Failed to retrieve Git status: %v", err)
			// GetStatusShort already logged details
			return err
		}

		// --- Present the Status Output ---
		trimmedStdout := strings.TrimSpace(stdout)
		if trimmedStdout == "" {
			presenter.Info("Working tree is clean.") // Use Info for clean status
			logger.InfoContext(ctx, "Status check reported clean working tree", slog.String("source_command", "status"))
		} else {
			// Use the Info block to display the short status lines
			presenter.InfoPrefixOnly()                                           // Print "INFO:" prefix
			fmt.Fprintln(presenter.Out(), "  Current Changes (--short format):") // Add context header
			scanner := bufio.NewScanner(strings.NewReader(trimmedStdout))
			for scanner.Scan() {
				// Print each line indented under the INFO block
				fmt.Fprintf(presenter.Out(), "    %s\n", scanner.Text())
			}
			presenter.Newline() // Add newline after the block
			logger.InfoContext(ctx, "Status check reported changes", slog.String("source_command", "status"), slog.Int("line_count", strings.Count(trimmedStdout, "\n")+1))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

======== END FILE: cmd/status.go ========

======== FILE: cmd/sync.go ========

// cmd/sync.go

package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	// "strings" // No longer needed directly here

	"github.com/contextvibes/cli/internal/git" // Use GitClient
	"github.com/contextvibes/cli/internal/ui"  // Use Presenter
	// "github.com/contextvibes/cli/internal/tools" // No longer needed for Git/Prompts
	"github.com/spf13/cobra"
)

// assumeYes defined in root.go

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs local branch with remote: ensures clean, pulls rebase, pushes if ahead.",
	Long: `Synchronizes the current local branch with its upstream remote counterpart.

Workflow:
1. Checks if the working directory is clean (no staged, unstaged, or untracked changes). Fails if dirty.
2. Determines current branch and remote.
3. Explains the plan (pull rebase, then push if needed).
4. Prompts for confirmation unless -y/--yes is specified.
5. Executes 'git pull --rebase'. Fails on conflicts or errors.
6. Checks if the local branch is ahead of the remote after the pull.
7. Executes 'git push' only if the branch was determined to be ahead.`,
	Example: `  contextvibes commit -m "Save work"  # Commit changes first if needed
  contextvibes sync                    # Sync the current branch
  contextvibes sync -y                 # Sync without confirmation`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Synchronizing local branch with remote.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Sync: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger} // Use defaults for remote/main branch from config if needed later
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "sync"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "sync"))

		// --- Check Prerequisites ---
		presenter.Info("Checking working directory status...")
		isClean, err := client.IsWorkingDirClean(ctx)
		if err != nil {
			presenter.Error("Failed to check working directory status: %v", err)
			return err // Client logs details
		}
		if !isClean {
			errMsg := "Working directory has uncommitted changes (staged, unstaged, or untracked)."
			presenter.Error(errMsg)
			presenter.Advice("Please commit or stash your changes before syncing. Try `contextvibes commit -m \"...\"`.")
			logger.WarnContext(ctx, "Sync prerequisite failed: working directory not clean", slog.String("source_command", "sync"))
			return errors.New("working directory not clean") // Use specific error
		}
		presenter.Info("Working directory is clean.")

		currentBranch, err := client.GetCurrentBranchName(ctx)
		if err != nil {
			// Less critical if we can't get the name for display, but log it.
			presenter.Warning("Could not determine current branch name: %v", err)
			currentBranch = "current branch" // Use placeholder for messages
		}
		remoteName := client.RemoteName() // Get configured remote name

		// --- Confirmation ---
		presenter.Newline()
		presenter.Info("Proposed Sync Actions:")
		presenter.Detail("1. Update local branch '%s' from remote '%s' (git pull --rebase).", currentBranch, remoteName)
		presenter.Detail("2. Push local changes to remote '%s' if local branch is ahead after update (git push).", remoteName)
		presenter.Newline()

		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(ctx, "Confirmation bypassed via flag", slog.String("source_command", "sync"), slog.Bool("yes_flag", true))
			confirmed = true
		} else {
			var promptErr error
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with sync?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading sync confirmation", slog.String("source_command", "sync"), slog.String("error", promptErr.Error()))
				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Sync aborted by user.")
			logger.InfoContext(ctx, "Sync aborted by user confirmation", slog.String("source_command", "sync"), slog.Bool("confirmed", false))
			return nil
		}
		logger.DebugContext(ctx, "Proceeding after sync confirmation", slog.String("source_command", "sync"), slog.Bool("confirmed", true))

		// --- Execute Sync ---
		presenter.Newline()
		presenter.Info("Step 1: Updating local branch '%s' from '%s'...", currentBranch, remoteName)
		// Note: PullRebase uses runGit, which pipes output. User will see git's output directly.
		if err := client.PullRebase(ctx, currentBranch); err != nil {
			presenter.Error("Error during 'git pull --rebase'. Resolve conflicts manually and then run 'contextvibes sync' again if needed.")
			// Client logs details
			// Return specific error from PullRebase
			return err
		}
		presenter.Info("Pull --rebase successful.") // User info

		presenter.Newline()
		presenter.Info("Step 2: Checking if push is needed...")
		isAhead, err := client.IsBranchAhead(ctx)
		if err != nil {
			// This is more serious, as we can't determine push status
			presenter.Error("Failed to check if branch is ahead of remote: %v", err)
			// Client logs details
			return err
		}

		if !isAhead {
			presenter.Info("Local branch '%s' is not ahead of remote '%s'. Push is not required.", currentBranch, remoteName)
			logger.InfoContext(ctx, "Push not required after pull", slog.String("source_command", "sync"))
		} else {
			presenter.Info("Local branch '%s' is ahead of remote '%s'. Pushing changes...", currentBranch, remoteName)
			logger.DebugContext(ctx, "Attempting push via client.Push", slog.String("source_command", "sync"))
			// Note: Push uses runGit, which pipes output. User will see git's output.
			if err := client.Push(ctx, currentBranch); err != nil {
				// Push method handles "up-to-date" gracefully, only real errors are returned
				presenter.Error("Error during 'git push': %v", err)
				// Client logs details
				return err
			}
			presenter.Info("Push successful.") // User info
		}

		presenter.Newline()
		presenter.Success("Sync completed successfully.") // Use Success
		logger.InfoContext(ctx, "Sync successful", slog.String("source_command", "sync"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

======== END FILE: cmd/sync.go ========

======== FILE: cmd/test.go ========

// cmd/test.go
package cmd

import (
	"context" // Ensure this is imported
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
	// "github.com/contextvibes/cli/internal/tools" // Should no longer be needed
)

// Define an interface matching the methods used by the helpers below.
// This makes the helpers testable independently of the global ExecClient.
type execTestClientInterface interface {
	CommandExists(commandName string) bool
	Execute(ctx context.Context, dir string, commandName string, args ...string) error
	// CaptureOutput might be needed if specific test runners' output needs parsing (not used by current test helpers)
}

var testCmd = &cobra.Command{
	Use:   "test [args...]",
	Short: "Runs project-specific tests (e.g., go test, pytest).",
	Long: `Detects the project type (Go, Python) and runs the appropriate test command.
Any arguments passed to 'contextvibes test' will be forwarded to the underlying test runner.

- Go: Runs 'go test ./...'
- Python: Runs 'pytest' (if available). Falls back to 'python -m unittest discover' if pytest not found.

For other project types, or if no specific test runner is found, it will indicate no action.`,
	Example: `  contextvibes test
  contextvibes test -v  # Passes '-v' to 'go test' or 'pytest'
  contextvibes test tests/my_specific_test.py # Passes path to pytest`,
	// Args: cobra.ArbitraryArgs, // Keep commented out unless strictly needed and understood
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger    // From cmd/root.go
		if ExecClient == nil { // From cmd/root.go
			return fmt.Errorf("internal error: executor client not initialized")
		}
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr(), os.Stdin)
		ctx := context.Background() // Get context

		presenter.Summary("Running project tests.")

		cwd, err := os.Getwd()
		if err != nil {
			wrappedErr := fmt.Errorf("failed to get current working directory: %w", err)
			logger.ErrorContext(ctx, "Test: Failed getwd", slog.String("error", err.Error()))
			presenter.Error("Failed to get current working directory: %v", err)
			return wrappedErr
		}

		presenter.Info("Detecting project type...")
		projType, err := project.Detect(cwd)
		if err != nil {
			wrappedErr := fmt.Errorf("failed to detect project type: %w", err)
			logger.ErrorContext(ctx, "Test: Failed project detection", slog.String("error", err.Error()))
			presenter.Error("Failed to detect project type: %v", err)
			return wrappedErr
		}
		presenter.Info("Detected project type: %s", presenter.Highlight(string(projType)))
		logger.Info("Project detection result", slog.String("source_command", "test"), slog.String("type", string(projType)))

		var testErr error
		testExecuted := false

		switch projType {
		case project.Go:
			presenter.Header("Go Project Tests")
			// The codemod should have changed this call to include ExecClient
			testErr = executeGoTests(ctx, presenter, logger, ExecClient, cwd, args)
			testExecuted = true
		case project.Python:
			presenter.Header("Python Project Tests")
			// The codemod should have changed this call to include ExecClient
			testErr = executePythonTests(ctx, presenter, logger, ExecClient, cwd, args)
			testExecuted = true
		case project.Terraform, project.Pulumi:
			presenter.Info("Automated testing for %s projects is not yet implemented in this command.", projType)
			presenter.Advice("Consider using tools like Terratest or language-specific test frameworks manually.")
		case project.Unknown:
			presenter.Warning("Unknown project type. Cannot determine how to run tests.")
		default:
			presenter.Warning("No specific test execution logic for project type: %s", projType)
		}

		presenter.Newline()
		if !testExecuted && testErr == nil {
			presenter.Info("No tests were executed for the detected project type or configuration.")
			return nil
		}

		if testErr != nil {
			presenter.Error("Project tests failed.")
			logger.Error("Test command finished with errors", slog.String("source_command", "test"), slog.String("error", testErr.Error()))
			return testErr
		}

		presenter.Success("Project tests completed successfully.")
		logger.Info("Test command successful", slog.String("source_command", "test"))
		return nil
	},
}

// Manually updated signature: accepts ctx, execClient
func executeGoTests(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execTestClientInterface, dir string, passThroughArgs []string) error {
	tool := "go"
	// This should have been updated by codemod
	if !execClient.CommandExists(tool) {
		errMsgForUser := fmt.Sprintf("'%s' command not found. Ensure Go is installed and in your PATH.", tool)
		presenter.Error(errMsgForUser)
		logger.Error("Go test prerequisite failed", slog.String("reason", errMsgForUser), slog.String("tool", tool))
		return errors.New("go command not found")
	}

	testArgs := []string{"test", "./..."}
	testArgs = append(testArgs, passThroughArgs...)

	presenter.Info("Executing: %s %s", tool, strings.Join(testArgs, " "))
	logger.Info("Executing go test", slog.String("source_command", "test"), slog.String("tool", tool), slog.Any("args", testArgs))

	// This should have been updated by codemod
	err := execClient.Execute(ctx, dir, tool, testArgs...)
	if err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}
	return nil
}

// Manually updated signature: accepts ctx, execClient
func executePythonTests(ctx context.Context, presenter *ui.Presenter, logger *slog.Logger, execClient execTestClientInterface, dir string, passThroughArgs []string) error {
	pytestTool := "pytest"
	pythonTool := "python"

	// This should have been updated by codemod
	if execClient.CommandExists(pytestTool) {
		presenter.Info("Executing: %s %s", pytestTool, strings.Join(passThroughArgs, " "))
		logger.Info("Executing pytest", slog.String("source_command", "test"), slog.String("tool", pytestTool), slog.Any("args", passThroughArgs))
		// This should have been updated by codemod
		err := execClient.Execute(ctx, dir, pytestTool, passThroughArgs...)
		if err != nil {
			return fmt.Errorf("pytest failed: %w", err)
		}
		return nil
	}

	presenter.Info("`pytest` not found. Attempting `python -m unittest discover`...")
	// This should have been updated by codemod
	if execClient.CommandExists(pythonTool) {
		unittestArgs := []string{"-m", "unittest", "discover"}

		presenter.Info("Executing: %s %s", pythonTool, strings.Join(unittestArgs, " "))
		logger.Info("Executing python unittest", slog.String("source_command", "test"), slog.String("tool", pythonTool), slog.Any("args", unittestArgs))
		// This should have been updated by codemod
		err := execClient.Execute(ctx, dir, pythonTool, unittestArgs...)
		if err != nil {
			return fmt.Errorf("python -m unittest discover failed: %w", err)
		}
		return nil
	}

	errMsgForUser := "Neither `pytest` nor `python` found. Cannot run Python tests."
	presenter.Error(errMsgForUser)
	logger.Error("Python test prerequisite failed", slog.String("reason", errMsgForUser))
	return errors.New("no python test runner found")
}

func init() {
	rootCmd.AddCommand(testCmd)
}

======== END FILE: cmd/test.go ========

======== FILE: cmd/version.go ========

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/contextvibes/cli/internal/ui"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the Context Vibes CLI",
	Long:  `Display the version number of the Context Vibes CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use cmd.OutOrStdout() so that output can be captured in tests
		p := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr(), os.Stdin)
		p.Summary(fmt.Sprintf("Context Vibes CLI Version: %s", AppVersion))
		return nil
	},
}

// init is called after all the variable declarations in the package have evaluated
// their initializers, and after all imported packages have been initialized.
// It is used here to add the versionCmd to the rootCmd.
// The AppVersion variable is expected to be initialized in root.go's init().
func init() {
	rootCmd.AddCommand(versionCmd)
}

======== END FILE: cmd/version.go ========

======== FILE: cmd/version_test.go ========

package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	// AppVersion is initialized in root.go's init() function.
	// For isolated testing, ensure it reflects the expected version if necessary,
	// but typically package-level init() should handle this.
	// If there were issues with init order across files (unlikely for same package),
	// one might set AppVersion = "0.1.0" here for explicit test control.

	testCases := []struct {
		name           string
		appVersion     string // To explicitly set for the test run
		expectedOutput string
	}{
		{
			name:           "Check Version Output",
			appVersion:     "0.1.0", // Ensure this matches what the test expects
			expectedOutput: "SUMMARY:\n  Context Vibes CLI Version: 0.1.0\n\n",
		},
		{
			name:           "Check Different Version Output",
			appVersion:     "v0.0.2-beta",
			expectedOutput: "SUMMARY:\n  Context Vibes CLI Version: v0.0.2-beta\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the global AppVersion for this specific test case run
			originalAppVersion := AppVersion // Store original to restore later
			AppVersion = tc.appVersion
			defer func() { AppVersion = originalAppVersion }() // Restore original AppVersion

			// Create a new root command for each test case to ensure isolation
			// and to properly wire up our versionCmd.
			// The actual rootCmd from root.go is not used directly to avoid init complexities
			// with logging or other persistent flags in a minimal test setup.
			testRootCmd := &cobra.Command{Use: "contextvibes"}

			// The versionCmd's init() function (which adds it to the global rootCmd)
			// will have run. To test it in isolation, we create a new root
			// and add the versionCmd to *it*.
			// Re-create versionCmd or ensure it's clean if it has state.
			// For this simple version command, just adding the package-level versionCmd is fine.
			testRootCmd.AddCommand(versionCmd)

			out := new(bytes.Buffer)
			testRootCmd.SetOut(out) // Capture standard output
			testRootCmd.SetErr(out) // Optionally capture stderr if needed, though version writes to stdout

			testRootCmd.SetArgs([]string{"version"})

			err := testRootCmd.Execute()
			if err != nil {
				t.Fatalf("Execute() failed unexpectedly: %v", err)
			}
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}

======== END FILE: cmd/version_test.go ========

======== FILE: cmd/wrapup.go ========

// cmd/wrapup.go

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	// "os/exec" // No longer needed

	"github.com/contextvibes/cli/internal/git"
	"github.com/contextvibes/cli/internal/ui"
	// "github.com/contextvibes/cli/internal/tools" // No longer needed
	"github.com/spf13/cobra"
)

// assumeYes defined in root.go

var wrapupCmd = &cobra.Command{
	Use:   "wrapup",
	Short: "Finalizes daily work: stages, commits (default msg), pushes.",
	Long: `Performs the end-of-day workflow: checks for local changes or if the branch
is ahead of the remote, stages changes, commits with a standard message if needed,
and pushes the current branch.

Requires confirmation unless -y/--yes is specified.`,
	Example: `  contextvibes wrapup   # Checks state, stages, commits (if needed), pushes, after confirmation
  contextvibes wrapup -y # Performs wrapup without confirmation`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := AppLogger
		if logger == nil {
			return fmt.Errorf("internal error: logger not initialized")
		}
		presenter := ui.NewPresenter(os.Stdout, os.Stderr, os.Stdin)
		ctx := context.Background()

		presenter.Summary("Starting end-of-day wrapup process.")

		// --- Init Git Client ---
		workDir, err := os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "Wrapup: Failed getwd", slog.String("err", err.Error()))
			presenter.Error("Failed getwd: %v", err)
			return err
		}
		gitCfg := git.GitClientConfig{Logger: logger}
		logger.DebugContext(ctx, "Initializing GitClient", slog.String("source_command", "wrapup"))
		client, err := git.NewClient(ctx, workDir, gitCfg)
		if err != nil {
			presenter.Error("Failed git init: %v", err)
			return err
		}
		logger.DebugContext(ctx, "GitClient initialized", slog.String("source_command", "wrapup"))

		// --- Check Repository State ---
		presenter.Info("Checking repository state...")
		isClean, err := client.IsWorkingDirClean(ctx)
		if err != nil {
			presenter.Error("Failed checking working directory status: %v", err)
			return err // Client logs details
		}

		isAhead := false // Assume not ahead initially
		if isClean {
			presenter.Info("Working directory is clean.")
			// Only check if ahead if clean
			isAhead, err = client.IsBranchAhead(ctx)
			if err != nil {
				// Log the error but proceed; assume push might still be needed if we commit later
				logger.WarnContext(ctx, "Could not determine if branch is ahead", slog.String("source_command", "wrapup"), slog.String("error", err.Error()))
				presenter.Warning("Could not accurately determine push status: %v", err)
				// We will still propose push if a commit happens
			} else {
				if isAhead {
					presenter.Info("Local branch is ahead of remote.")
				} else {
					presenter.Info("Local branch is not ahead of remote.")
				}
			}
		} else {
			presenter.Info("Changes detected in working directory.")
			// If dirty, we'll commit, so push is definitely intended.
			isAhead = true // Treat as ahead for planning purposes if dirty
		}

		// --- Determine Actions ---
		actionCommit := !isClean              // Only commit if initially dirty
		actionPush := isAhead || actionCommit // Push if already ahead OR if we are about to commit

		commitMsg := "chore: Automated wrapup commit" // Default message

		if !actionCommit && !actionPush {
			presenter.Newline()
			presenter.Success("No actions needed (no local changes to commit and branch is not ahead).")
			logger.InfoContext(ctx, "Wrapup complete: no actions needed", slog.String("source_command", "wrapup"))
			// Add workflow advice even if no actions were needed (moved here for consistency)
			presenter.Newline()
			presenter.Advice("`wrapup` is an automated shortcut. For quality checks, custom commits, and pre-push sync, consider running:")
			presenter.Advice("  `contextvibes quality && contextvibes commit -m '...' && contextvibes sync`")
			return nil
		}

		// --- Confirmation ---
		presenter.Newline()
		presenter.Info("Proposed Wrapup Actions:")
		actionCounter := 1
		if actionCommit {
			presenter.Detail("%d. Stage all changes (git add .)", actionCounter)
			actionCounter++
			presenter.Detail("%d. Commit staged changes with message: '%s'", actionCounter, commitMsg)
			actionCounter++
		}
		if actionPush {
			presenter.Detail("%d. Push current branch to remote '%s'", actionCounter, client.RemoteName())
		}
		presenter.Newline()

		// *** MOVED ADVICE HERE ***
		presenter.Advice("`wrapup` is an automated shortcut. For quality checks, custom commits, and pre-push sync, consider running:")
		presenter.Advice("  `contextvibes quality && contextvibes commit -m '...' && contextvibes sync`")
		presenter.Newline() // Add space before the prompt
		// ***********************

		confirmed := false
		if assumeYes {
			presenter.Info("Confirmation prompt bypassed via --yes flag.")
			logger.InfoContext(ctx, "Confirmation bypassed via flag", slog.String("source_command", "wrapup"), slog.Bool("yes_flag", true))
			confirmed = true
		} else {
			var promptErr error
			confirmed, promptErr = presenter.PromptForConfirmation("Proceed with wrapup?")
			if promptErr != nil {
				logger.ErrorContext(ctx, "Error reading wrapup confirmation", slog.String("source_command", "wrapup"), slog.String("error", promptErr.Error()))
				return promptErr
			}
		}

		if !confirmed {
			presenter.Info("Wrapup aborted by user.")
			logger.InfoContext(ctx, "Wrapup aborted by user confirmation", slog.String("source_command", "wrapup"), slog.Bool("confirmed", false))
			return nil
		}
		logger.DebugContext(ctx, "Proceeding after wrapup confirmation", slog.String("source_command", "wrapup"), slog.Bool("confirmed", true))

		// --- Execute Actions ---
		presenter.Newline()
		commitActuallyHappened := false
		if actionCommit {
			presenter.Info("Staging all changes...")
			logger.DebugContext(ctx, "Attempting stage via client.AddAll", slog.String("source_command", "wrapup"))
			if err := client.AddAll(ctx); err != nil {
				presenter.Error("Failed to stage changes: %v", err)
				return err
			}

			// Check if staging actually resulted in changes to be committed
			logger.DebugContext(ctx, "Checking staged status after add", slog.String("source_command", "wrapup"))
			commitIsNeeded, err := client.HasStagedChanges(ctx)
			if err != nil {
				presenter.Error("Failed to check staged status after add: %v", err)
				return err
			}

			if commitIsNeeded {
				presenter.Info("Committing staged changes with message: '%s'...", commitMsg)
				logger.DebugContext(ctx, "Attempting commit via client.Commit", slog.String("source_command", "wrapup"))
				if err := client.Commit(ctx, commitMsg); err != nil {
					presenter.Error("Failed to commit changes: %v", err)
					return err
				}
				commitActuallyHappened = true
			} else {
				presenter.Info("No changes were staged after 'git add .', skipping commit step.")
				logger.InfoContext(ctx, "Commit skipped, no changes staged", slog.String("source_command", "wrapup"))
			}
		}

		if actionPush {
			// Determine branch to push (usually the current one)
			// GetCurrentBranchName is relatively safe even if called again
			branchToPush, branchErr := client.GetCurrentBranchName(ctx)
			if branchErr != nil {
				// If we can't get branch name now, it's a problem for push
				presenter.Error("Cannot determine current branch to push: %v", branchErr)
				return branchErr
			}

			presenter.Info("Pushing branch '%s' to remote '%s'...", branchToPush, client.RemoteName())
			logger.DebugContext(ctx, "Attempting push via client.Push", slog.String("source_command", "wrapup"), slog.String("branch", branchToPush))
			if err := client.Push(ctx, branchToPush); err != nil {
				// client.Push handles "up-to-date" logging internally and returns nil for it.
				// Only real errors should be returned here.
				presenter.Error("Failed to push changes: %v", err)
				return err
			}
			presenter.Info("Push successful or already up-to-date.")
		} else {
			presenter.Info("Skipping push step as no push action was planned.")
			logger.InfoContext(ctx, "Push skipped (not needed)", slog.String("source_command", "wrapup"))
		}

		// --- Final Status ---
		presenter.Newline()
		presenter.Success("Wrapup complete.")
		// Removed the advice from here as it's now shown before confirmation.
		// Just state what happened.
		if commitActuallyHappened && actionPush {
			presenter.Detail("Local changes were committed and the branch was pushed.")
		} else if commitActuallyHappened {
			presenter.Detail("Local changes were committed (push was not needed or skipped).")
		} else if actionPush {
			presenter.Detail("Branch was pushed (no local changes needed committing).")
		}

		logger.InfoContext(ctx, "Wrapup successful",
			slog.String("source_command", "wrapup"),
			slog.Bool("commit_executed", commitActuallyHappened),
			slog.Bool("push_executed", actionPush),
		)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(wrapupCmd)
}

======== END FILE: cmd/wrapup.go ========

======== FILE: docs/COMMAND_REFERENCE.md ========

## Command Reference (Reference)

This section provides a detailed reference for each command in the Context Vibes CLI.

### `describe`

**Synopsis:**

```
contextvibes describe [-o <output_file>]
```

**Description:**

Gathers project context (user prompt, environment, git status, structure, relevant files) and writes it to a Markdown file, suitable for AI interaction. The default output file is `contextvibes.md`.

**Flags:**

| Flag          | Short | Description                                                              | Data Type | Default Value      | Overrides Config File |
|---------------|-------|--------------------------------------------------------------------------|-----------|--------------------|-----------------------|
| `--output`    | `-o`  | Path to write the context markdown file.                                 | string    | `contextvibes.md`  | No                    |

**Example Usage:**

*   Generate a context file with the default name:

    ```bash
    contextvibes describe
    ```

*   Generate a context file with a custom name:

    ```bash
    contextvibes describe -o project_snapshot.md
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                   |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The context file was generated successfully.                                                                                   |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: empty prompt, read failures. |

### `diff`

**Synopsis:**

```
contextvibes diff
```

**Description:**

Generates a Markdown summary of pending Git changes (staged, unstaged, untracked) in the Git repository and **overwrites** the context file: `contextvibes.md`.

**Flags:**

This command has no flags.

**Example Usage:**

```bash
contextvibes diff
```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                   |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The diff summary was generated successfully, or no changes were found.                                                                                   |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: Git command failures. |

### `kickoff`

**Synopsis:**

```
contextvibes kickoff [--branch <branch-name>]
```

**Description:**

Performs the start-of-work workflow: updates the main branch from the remote, creates and switches to a new branch, and pushes the new branch to the remote. Requires a clean state on the main branch.

**Flags:**

| Flag       | Short | Description                                                                   | Data Type | Default Value | Overrides Config File |
|------------|-------|-------------------------------------------------------------------------------|-----------|---------------|-----------------------|
| `--branch` | `-b`  | Name for the new branch (e.g., `feature/JIRA-123-task-name`). If omitted, user is prompted. | string    | ""            | No                    |

**Example Usage:**

*   Start a new feature branch with a specified name:

    ```bash
    contextvibes kickoff --branch feature/JIRA-123-new-widget
    ```

*   Start a new fix branch with a specified name, bypassing confirmation:

    ```bash
    contextvibes kickoff -b fix/login-bug -y
    ```

*   Start a new branch, prompting for the name:

    ```bash
    contextvibes kickoff
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                                                                               |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The new branch was created and pushed successfully.                                                                                                                                                        |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: dirty working directory, not on main branch, invalid branch name, remote access issues, etc. |

### `commit`

**Synopsis:**

```
contextvibes commit -m <message>
```

**Description:**

Stages all changes and commits locally with a provided message.  Commit message validation is active by default, and the rules are configurable via `.contextvibes.yaml`.

**Flags:**

| Flag        | Short | Description                         | Data Type | Default Value | Overrides Config File |
|-------------|-------|-------------------------------------|-----------|---------------|-----------------------|
| `--message` | `-m`  | Commit message (required).          | string    | ""            | No                    |

**Example Usage:**

*   Commit changes with a message:

    ```bash
    contextvibes commit -m "feat(auth): Implement OTP login"
    ```

*   Commit changes with a message, bypassing confirmation:

    ```bash
    contextvibes commit -m "fix(api): Correct typo in user model" -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                                                                     |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The changes were staged and committed successfully.                                                                                                                                               |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: missing commit message, invalid commit message format, Git command failures, etc. |

### `sync`

**Synopsis:**

```
contextvibes sync
```

**Description:**

Syncs the local branch with the remote, ensuring it's clean, pulling with rebase, and pushing if ahead.

**Flags:**

This command has no flags.

**Example Usage:**

*   Sync the current branch:

    ```bash
    contextvibes sync
    ```

*   Sync the current branch, bypassing confirmation:

    ```bash
    contextvibes sync -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                           |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The branch was synced successfully.                                                                                                                    |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: dirty working directory, pull rebase failures, remote access issues.  |

### `wrapup`

**Synopsis:**

```
contextvibes wrapup
```

**Description:**

Finalizes daily work: stages all changes, commits (with a default message if needed), and pushes the current branch.

**Flags:**

This command has no flags.

**Example Usage:**

*   Wrap up the current branch:

    ```bash
    contextvibes wrapup
    ```

*   Wrap up the current branch, bypassing confirmation:

    ```bash
    contextvibes wrapup -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                              |
|-----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The changes were staged, committed (if needed), and the branch was pushed (if needed).                                                                     |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: git failures, merge conflicts, or remote errors. |

### `plan`

**Synopsis:**

```
contextvibes plan
```

**Description:**

Generates an execution plan (e.g., terraform plan, pulumi preview) based on the detected project type.

**Flags:**

This command has no flags.

**Example Usage:**

*   Generate a plan:

    ```bash
    contextvibes plan
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                               |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success (no changes detected).  Terraform: no changes detected. Pulumi: preview completed successfully.                            |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: missing tools, invalid configuration files.   |
| 2         | Terraform only: plan indicates changes are needed (success for plan command itself).                                                   |

### `deploy`

**Synopsis:**

```
contextvibes deploy
```

**Description:**

Deploys infrastructure changes (terraform apply, pulumi up) based on the detected project type.

**Flags:**

This command has no flags.

**Example Usage:**

*   Deploy changes:

    ```bash
    contextvibes deploy
    ```

* Deploy changes automatically:

    ```bash
    contextvibes deploy -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                  |
|-----------|--------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. Deployment completed successfully.                                                                            |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Missing tfplan.out.  |

### `quality`

**Synopsis:**

```
contextvibes quality
```

**Description:**

Runs code formatting and linting checks (Terraform, Python, Go) based on the detected project type.

**Flags:**

This command has no flags.

**Example Usage:**

*   Run quality checks:

    ```bash
    contextvibes quality
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                               |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. All quality checks passed successfully, or warnings were reported but no critical errors occurred.                                         |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. Common causes:  formatting violations, linter errors.  |

### `format`

**Synopsis:**

```
contextvibes format
```

**Description:**

Applies code formatting (go fmt, terraform fmt, isort, black) based on the detected project type.

**Flags:**

This command has no flags.

**Example Usage:**

*   Apply code formatting:

    ```bash
    contextvibes format
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                       |
|-----------|-----------------------------------------------------------------------------------------------|
| 0         | Success. All formatting tools completed successfully or applied changes.                       |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. |

### `test`

**Synopsis:**

```
contextvibes test [args...]
```

**Description:**

Runs project-specific tests (e.g., go test, pytest) based on the detected project type. Any arguments are passed to the test runner.

**Flags:**

This command accepts arbitrary arguments passed to the underlying test runner.

**Example Usage:**

*   Run tests:

    ```bash
    contextvibes test
    ```

* Run verbose go tests:

    ```bash
    contextvibes test -v
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                     |
|-----------|---------------------------------------------------------------------------------------------|
| 0         | Success. All tests passed successfully.                                                     |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. |

### `version`

**Synopsis:**

```
contextvibes version
```

**Description:**

Displays the version number of the Context Vibes CLI.

**Flags:**

This command has no flags.

**Example Usage:**

*   Display the version:

    ```bash
    contextvibes version
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                    |
|-----------|----------------------------------------------------------------------------|
| 0         | Success. The version number was displayed successfully.                   |
| 1         | An error occurred. Check the terminal and AI log file, though this is unlikely.  |

### `codemod`

**Synopsis:**

```
contextvibes codemod [--script <file.json>]
```

**Description:**

Applies programmatic code modifications or deletions from a JSON script.

**Flags:**

| Flag       | Short | Description                                              | Data Type | Default Value | Overrides Config File |
|------------|-------|----------------------------------------------------------|-----------|---------------|-----------------------|
| `--script` | `-s`  | Path to the JSON codemod script file.                  | string    | `codemod.json`| No                    |

**Example Usage:**

*   Run the codemod using the default script:

    ```bash
    contextvibes codemod
    ```

*   Run the codemod using a custom script:

    ```bash
    contextvibes codemod --script ./my_refactor_script.json
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                       |
|-----------|-----------------------------------------------------------------------------------------------|
| 0         | Success. The codemod script was executed successfully.                                         |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. |

======== END FILE: docs/COMMAND_REFERENCE.md ========

======== FILE: docs/CONFIGURATION_REFERENCE.md ========

## Configuration Reference (Reference)

The Context Vibes CLI can be configured using a `.contextvibes.yaml` file located in the root directory of your project. This file allows you to customize various aspects of the CLI's behavior. If the file is not present, the CLI uses sensible built-in defaults.

### File Format

The `.contextvibes.yaml` file uses YAML syntax.

### Top-Level Sections

The configuration file is divided into the following top-level sections:

*   `git`: Settings related to Git repository interaction.
*   `logging`: Settings related to logging.
*   `validation`: Settings related to input validation.

### Section Details

#### `git`

This section configures Git-related settings.

| Key                  | Data Type | Description                                                                                                                      | Default Value |
|----------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------|---------------|
| `defaultRemote`      | string    | The name of the default Git remote.                                                                                              | `origin`        |
| `defaultMainBranch`  | string    | The name of the default main branch (used by `kickoff` and other commands).                                                      | `main`          |

Example:

```yaml
git:
  defaultRemote: origin
  defaultMainBranch: main```

#### `logging`

This section configures logging settings.

| Key               | Data Type | Description                                                                                                                               | Default Value             |
|-------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|---------------------------|
| `defaultAILogFile` | string    | The file path for the detailed AI JSON log. This setting is overridden by the `--ai-log-file` command-line flag, if provided.        | `contextvibes_ai_trace.log` |

Example:

```yaml
logging:
  defaultAILogFile: "logs/contextvibes_ai.log" # Path relative to project root
```

#### `validation`

This section configures input validation rules.

##### `validation.branchName`

| Key       | Data Type | Description                                                                                                                                          | Default Value                                    |
|-----------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------|
| `enable`  | boolean   | Whether branch name validation is enabled. If not set, defaults to `true`.                                                                            | `true`                                           |
| `pattern` | string    | A regular expression pattern used to validate branch names. If `enable` is `true` and `pattern` is not set, a default pattern is used.    | `^((feature|fix|docs|format)/.+)$`               |

##### `validation.commitMessage`

| Key       | Data Type | Description                                                                                                                                            | Default Value                                                                                                                                                                |
|-----------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `enable`  | boolean   | Whether commit message validation is enabled. If not set, defaults to `true`.                                                                            | `true`                                                                                                                                                                      |
| `pattern` | string    | A regular expression pattern used to validate commit messages. If `enable` is `true` and `pattern` is not set, a default Conventional Commits pattern is used. | `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+$` |

Example:

```yaml
validation:
  branchName:
    enable: true # Default is true. Set to false to disable validation.
    # Default pattern if enabled and not specified: ^((feature|fix|docs|format)/.+)$
    pattern: "^(feature|fix|chore|task)/[A-Z]+-[0-9]+-.*$" # Example custom pattern
  commitMessage:
    enable: true # Default is true. Set to false to disable validation.
    # Default pattern if enabled and not specified: ^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+
    pattern: "^(feat|fix|chore|docs)!?(\\(.+\\))?: .+" # Example custom pattern
```

### Precedence

The configuration settings are applied in the following order of precedence (highest to lowest):

1.  **Command-line flags:** Flags provided directly when running a command (e.g., `--ai-log-file`, `--log-level-ai`) always override any other settings.
2.  **`.contextvibes.yaml` file:** Settings defined in this file override the built-in defaults if the file exists and the setting is specified.
3.  **Built-in Defaults:** The default values hardcoded within the CLI application (defined in `internal/config/config.go`).

This means that if a setting is specified both in the configuration file and as a command-line flag, the command-line flag will take precedence. If no config file is found, or the setting isn't specified in the config file or via a flag, the built-in default value will be used.

======== END FILE: docs/CONFIGURATION_REFERENCE.md ========

======== FILE: go.mod ========

module github.com/contextvibes/cli

go 1.24.2

require (
	github.com/denormal/go-gitignore v0.0.0-20180930084346-ae8ad1d07817
	github.com/fatih/color v1.18.0
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

======== END FILE: go.mod ========

======== FILE: go.sum ========

github.com/cpuguy83/go-md2man/v2 v2.0.6/go.mod h1:oOW0eioCTA6cOiMLiUPZOpcVxMig6NIQQ7OS05n1F4g=
github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 h1:y5HC9v93H5EPKqaS1UYVg1uYah5Xf51mBfIoWehClUQ=
github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964/go.mod h1:Xd9hchkHSWYkEqJwUGisez3G1QY8Ryz0sdWrLPMGjLk=
github.com/davecgh/go-spew v1.1.1 h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=
github.com/davecgh/go-spew v1.1.1/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
github.com/denormal/go-gitignore v0.0.0-20180930084346-ae8ad1d07817 h1:0nsrg//Dc7xC74H/TZ5sYR8uk4UQRNjsw8zejqH5a4Q=
github.com/denormal/go-gitignore v0.0.0-20180930084346-ae8ad1d07817/go.mod h1:C/+sI4IFnEpCn6VQ3GIPEp+FrQnQw+YQP3+n+GdGq7o=
github.com/fatih/color v1.18.0 h1:S8gINlzdQ840/4pfAwic/ZE0djQEH3wM94VfqLTZcOM=
github.com/fatih/color v1.18.0/go.mod h1:4FelSpRwEGDpQ12mAdzqdOukCy4u8WUtOY6lkT/6HfU=
github.com/inconshreveable/mousetrap v1.1.0 h1:wN+x4NVGpMsO7ErUn/mUI3vEoE6Jt13X2s0bqwp9tc8=
github.com/inconshreveable/mousetrap v1.1.0/go.mod h1:vpF70FUmC8bwa3OWnCshd2FqLfsEA9PFc4w1p2J65bw=
github.com/mattn/go-colorable v0.1.13 h1:fFA4WZxdEF4tXPZVKMLwD8oUnCTTo08duU7wxecdEvA=
github.com/mattn/go-colorable v0.1.13/go.mod h1:7S9/ev0klgBDR4GtXTXX8a3vIGJpMovkB8vQcUbaXHg=
github.com/mattn/go-isatty v0.0.16/go.mod h1:kYGgaQfpe5nmfYZH+SKPsOc2e4SrIfOl2e/yFXSvRLM=
github.com/mattn/go-isatty v0.0.20 h1:xfD0iDuEKnDkl03q4limB+vH+GxLEtL/jb4xVJSWWEY=
github.com/mattn/go-isatty v0.0.20/go.mod h1:W+V8PltTTMOvKvAeJH7IuucS94S2C6jfK/D7dTCTo3Y=
github.com/pmezard/go-difflib v1.0.0 h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=
github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
github.com/russross/blackfriday/v2 v2.1.0/go.mod h1:+Rmxgy9KzJVeS9/2gXHxylqXiyQDYRxCVz55jmeOWTM=
github.com/spf13/cobra v1.9.1 h1:CXSaggrXdbHK9CF+8ywj8Amf7PBRmPCOJugH954Nnlo=
github.com/spf13/cobra v1.9.1/go.mod h1:nDyEzZ8ogv936Cinf6g1RU9MRY64Ir93oCnqb9wxYW0=
github.com/spf13/pflag v1.0.6 h1:jFzHGLGAlb3ruxLB8MhbI6A8+AQX/2eW4qeyNZXNp2o=
github.com/spf13/pflag v1.0.6/go.mod h1:McXfInJRrz4CZXVZOBLb0bTZqETkiAhM9Iw0y3An2Bg=
github.com/stretchr/testify v1.10.0 h1:Xv5erBjTwe/5IxqUQTdXv5kgmIvbHo3QQyRwhJsOfJA=
github.com/stretchr/testify v1.10.0/go.mod h1:r2ic/lqez/lEtzL7wO/rwa5dbSLXVDPFyf8C91i36aY=
golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.6.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.25.0 h1:r+8e+loiHxRqhXVl6ML1nO3l1+oFoWbnlu2Ehimmi34=
golang.org/x/sys v0.25.0/go.mod h1:/VUhepiaJMQUp4+oa/7Zr1D23ma6VTLIYjOOTFZPUcA=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405 h1:yhCVgyC4o1eVCa2tZl7eS0r+SDo693bJlVdllGtEeKM=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=

======== END FILE: go.sum ========

======== FILE: internal/codemod/doc.go ========

/*
Package codemod defines the data structures used to represent codification modification
scripts for the contextvibes CLI. These structures allow for a standardized way
to describe a series of automated changes to files within a codebase.

The core types are:
  - Operation: Defines a single modification to be performed on a file, such as
    a regular expression replacement or a file deletion. It includes fields like
    `Type`, `Description`, `FindRegex`, and `ReplaceWith`.
  - FileChangeSet: Groups all `Operation`s intended for a single target file,
    specified by `FilePath`.
  - ChangeScript: Represents the top-level structure of a codemod script, which is
    an array of `FileChangeSet`s, allowing modifications across multiple files.

These types are typically unmarshalled from a JSON file (e.g., the default
`contextvibes-codemod.json` or a user-specified script) by the
`contextvibes codemod` command. The command then interprets these structures
to apply the requested changes to the project's files.

This package itself does not contain the execution logic for applying the
codemods; that logic resides in the `cmd` package (specifically `cmd/codemod.go`).
The primary role of `internal/codemod` is to provide the clear, typed
representation of the modification instructions.
*/
package codemod

======== END FILE: internal/codemod/doc.go ========

======== FILE: internal/codemod/types.go ========

package codemod

// Operation defines a single modification to be performed on a file.
type Operation struct {
	// Type indicates the kind of operation (e.g., "regex_replace", "add_import").
	Type string `json:"type"`
	// Description provides a human-readable explanation of the operation.
	Description string `json:"description,omitempty"`

	// --- Fields for "regex_replace" type ---
	// FindRegex is the regular expression to find.
	FindRegex string `json:"find_regex,omitempty"`
	// ReplaceWith is the string to replace matches with.
	ReplaceWith string `json:"replace_with,omitempty"`
	// LineNumber can be used to target a specific line for some operations (not used by basic regex_replace yet).
	LineNumber *int `json:"line_number,omitempty"`

	// --- Fields for "delete_file" type ---
	// No specific fields needed for simple delete, relies on FileChangeSet.FilePath

	// Future operations might include:
	// For "add_import_if_missing":
	// ImportPath string `json:"import_path,omitempty"`

	// For "comment_update":
	// OldCommentRegex string `json:"old_comment_regex,omitempty"`
	// NewComment      string `json:"new_comment,omitempty"`
}

// FileChangeSet groups all operations for a single file.
type FileChangeSet struct {
	FilePath   string      `json:"file_path"`
	Operations []Operation `json:"operations"`
}

// ChangeScript is the top-level structure, representing a list of changes for multiple files.
type ChangeScript []FileChangeSet

======== END FILE: internal/codemod/types.go ========

======== FILE: internal/config/config.go ========

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec"
	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigFileName        = ".contextvibes.yaml"
	DefaultCodemodFilename       = "codemod.json"
	DefaultDescribeOutputFile    = "contextvibes.md"
	DefaultBranchNamePattern     = `^((feature|fix|docs|format)/.+)$`
	DefaultCommitMessagePattern  = `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+`
	DefaultGitRemote             = "origin"
	DefaultGitMainBranch         = "main"
	UltimateDefaultAILogFilename = "contextvibes_ai_trace.log"
)

type GitSettings struct {
	DefaultRemote     string `yaml:"defaultRemote,omitempty"`
	DefaultMainBranch string `yaml:"defaultMainBranch,omitempty"`
}

type ValidationRule struct {
	Enable  *bool  `yaml:"enable,omitempty"`
	Pattern string `yaml:"pattern,omitempty"`
}

type LoggingSettings struct {
	DefaultAILogFile string `yaml:"defaultAILogFile,omitempty"`
}

type ProjectState struct {
	StrategicKickoffCompleted *bool  `yaml:"strategicKickoffCompleted,omitempty"`
	LastStrategicKickoffDate  string `yaml:"lastStrategicKickoffDate,omitempty"`
}

type AICollaborationPreferences struct {
	CodeProvisioningStyle string `yaml:"codeProvisioningStyle,omitempty"`
	MarkdownDocsStyle     string `yaml:"markdownDocsStyle,omitempty"`
	DetailedTaskMode      string `yaml:"detailedTaskMode,omitempty"`
	ProactiveDetailLevel  string `yaml:"proactiveDetailLevel,omitempty"`
	AIProactivity         string `yaml:"aiProactivity,omitempty"`
}

type AISettings struct {
	CollaborationPreferences AICollaborationPreferences `yaml:"collaborationPreferences,omitempty"`
}

type Config struct {
	Git        GitSettings     `yaml:"git,omitempty"`
	Logging    LoggingSettings `yaml:"logging,omitempty"`
	Validation struct {
		BranchName    ValidationRule `yaml:"branchName,omitempty"`
		CommitMessage ValidationRule `yaml:"commitMessage,omitempty"`
	} `yaml:"validation,omitempty"`
	ProjectState ProjectState `yaml:"projectState,omitempty"`
	AI           AISettings   `yaml:"ai,omitempty"`
}

func GetDefaultConfig() *Config {
	enableTrue := true    
	defaultFalse := false 

	cfg := &Config{
		Git: GitSettings{
			DefaultRemote:     DefaultGitRemote,
			DefaultMainBranch: DefaultGitMainBranch,
		},
		Logging: LoggingSettings{
			DefaultAILogFile: UltimateDefaultAILogFilename,
		},
		Validation: struct {
			BranchName    ValidationRule `yaml:"branchName,omitempty"`
			CommitMessage ValidationRule `yaml:"commitMessage,omitempty"`
		}{
			BranchName: ValidationRule{
				Enable:  &enableTrue,
				Pattern: DefaultBranchNamePattern,
			},
			CommitMessage: ValidationRule{
				Enable:  &enableTrue,
				Pattern: DefaultCommitMessagePattern,
			},
		},
		ProjectState: ProjectState{ 
			StrategicKickoffCompleted: &defaultFalse, 
			LastStrategicKickoffDate:  "",
		},
		AI: AISettings{ 
			CollaborationPreferences: AICollaborationPreferences{},
		},
	}
	return cfg
}

func LoadConfig(filePath string) (*Config, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil 
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filePath, err)
	}
	if len(data) == 0 { 
		return nil, nil 
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", filePath, err)
	}
	return &cfg, nil
}

func FindRepoRootConfigPath(execClient *exec.ExecutorClient) (string, error) {
	if execClient == nil {
		return "", errors.New("executor client is nil, cannot find repo root")
	}
	ctx := context.Background() 
	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to determine git repository root (is this a git repo, or is 'git' not in PATH? details: %s): %w", strings.TrimSpace(stderr), err)
	}
	repoRoot := filepath.Clean(strings.TrimSpace(stdout))
	if repoRoot == "" || repoRoot == "." { 
		return "", errors.New("git rev-parse --show-toplevel returned an empty or invalid path, not in a git repository")
	}

	configPath := filepath.Join(repoRoot, DefaultConfigFileName)
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		return "", nil 
	} else if statErr != nil {
		return "", fmt.Errorf("error checking for config file at '%s': %w", configPath, statErr)
	}
	return configPath, nil
}

func MergeWithDefaults(loadedCfg *Config, defaultConfig *Config) *Config {
	if defaultConfig == nil { 
		panic("MergeWithDefaults called with nil defaultConfig")
	}
	if loadedCfg == nil {
		return defaultConfig 
	}

	finalCfg := *defaultConfig 

	if loadedCfg.Git.DefaultRemote != "" {
		finalCfg.Git.DefaultRemote = loadedCfg.Git.DefaultRemote
	}
	if loadedCfg.Git.DefaultMainBranch != "" {
		finalCfg.Git.DefaultMainBranch = loadedCfg.Git.DefaultMainBranch
	}
	if loadedCfg.Logging.DefaultAILogFile != "" {
		finalCfg.Logging.DefaultAILogFile = loadedCfg.Logging.DefaultAILogFile
	}

	if loadedCfg.Validation.BranchName.Enable != nil {
		finalCfg.Validation.BranchName.Enable = loadedCfg.Validation.BranchName.Enable
	}
	if finalCfg.Validation.BranchName.Enable == nil || *finalCfg.Validation.BranchName.Enable {
		if loadedCfg.Validation.BranchName.Pattern != "" {
			finalCfg.Validation.BranchName.Pattern = loadedCfg.Validation.BranchName.Pattern
		}
	} else { 
		finalCfg.Validation.BranchName.Pattern = "" 
	}

	if loadedCfg.Validation.CommitMessage.Enable != nil {
		finalCfg.Validation.CommitMessage.Enable = loadedCfg.Validation.CommitMessage.Enable
	}
	if finalCfg.Validation.CommitMessage.Enable == nil || *finalCfg.Validation.CommitMessage.Enable {
		if loadedCfg.Validation.CommitMessage.Pattern != "" {
			finalCfg.Validation.CommitMessage.Pattern = loadedCfg.Validation.CommitMessage.Pattern
		}
	} else {
		finalCfg.Validation.CommitMessage.Pattern = ""
	}

	if loadedCfg.ProjectState.StrategicKickoffCompleted != nil { 
		finalCfg.ProjectState.StrategicKickoffCompleted = loadedCfg.ProjectState.StrategicKickoffCompleted
	}
	if loadedCfg.ProjectState.LastStrategicKickoffDate != "" {
		finalCfg.ProjectState.LastStrategicKickoffDate = loadedCfg.ProjectState.LastStrategicKickoffDate
	}
	
	userAICollabPrefs := loadedCfg.AI.CollaborationPreferences
	if userAICollabPrefs.CodeProvisioningStyle != "" {
		finalCfg.AI.CollaborationPreferences.CodeProvisioningStyle = userAICollabPrefs.CodeProvisioningStyle
	}
	if userAICollabPrefs.MarkdownDocsStyle != "" {
		finalCfg.AI.CollaborationPreferences.MarkdownDocsStyle = userAICollabPrefs.MarkdownDocsStyle
	}
	if userAICollabPrefs.DetailedTaskMode != "" {
		finalCfg.AI.CollaborationPreferences.DetailedTaskMode = userAICollabPrefs.DetailedTaskMode
	}
	if userAICollabPrefs.ProactiveDetailLevel != "" {
		finalCfg.AI.CollaborationPreferences.ProactiveDetailLevel = userAICollabPrefs.ProactiveDetailLevel
	}
	if userAICollabPrefs.AIProactivity != "" {
		finalCfg.AI.CollaborationPreferences.AIProactivity = userAICollabPrefs.AIProactivity
	}

	return &finalCfg
}

// UpdateAndSaveConfig marshals the provided Config struct to YAML and writes it to filePath.
// It attempts to perform an atomic write by writing to a temporary file first, then renaming.
// NOTE: This will overwrite the existing file and may strip comments/reorder fields due to yaml.Marshal.
func UpdateAndSaveConfig(cfgToSave *Config, filePath string) error {
	if cfgToSave == nil {
		return errors.New("cannot save a nil config to file")
	}

	yamlData, err := yaml.Marshal(cfgToSave)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML for saving: %w", err)
	}

	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for config file '%s': %w", dir, err)
		}
	}

	// Atomic write: write to temp file then rename
	tempFile, err := os.CreateTemp(dir, filepath.Base(filePath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file in '%s': %w", dir, err)
	}
	
	// Ensure temp file is removed if anything goes wrong before rename
	// or after a successful rename (though rename moves it).
	// If rename succeeds, os.Remove will fail on tempFile.Name(), which is fine.
	defer os.Remove(tempFile.Name()) 

	if _, err := tempFile.Write(yamlData); err != nil {
		// Attempt to close even on write error, then return the write error
		_ = tempFile.Close() // Best effort close, ignore error here as Write error is primary
		return fmt.Errorf("failed to write to temporary config file '%s': %w", tempFile.Name(), err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary config file '%s': %w", tempFile.Name(), err)
	}

	// Rename temporary file to the actual config file path
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("failed to rename temporary config file to '%s': %w", filePath, err)
	}

	return nil
}

======== END FILE: internal/config/config.go ========

======== FILE: internal/config/doc.go ========

/*
Package config manages the configuration for the contextvibes CLI application.

It defines the structure of the configuration, provides functions to load
configuration from a YAML file (defaulting to '.contextvibes.yaml' in the
project root), and offers a way to get default configuration values.

The primary components are:
  - Config: The main struct holding all configuration settings, including Git behavior,
    logging preferences, and validation rules for branch names and commit messages.
  - GitSettings, LoggingSettings, ValidationRule: Sub-structs organizing related
    configuration items.
  - GetDefaultConfig(): Returns a pointer to a Config struct populated with sensible
    default values.
  - LoadConfig(filePath string): Attempts to load configuration from the specified
    YAML file. Returns nil if the file doesn't exist, allowing graceful fallback to defaults.
  - FindRepoRootConfigPath(execClient *exec.ExecutorClient): Locates the configuration
    file by searching upwards from the current directory to the Git repository root.
  - MergeWithDefaults(loadedCfg *Config, defaultConfig *Config): Merges a loaded
    user configuration with the default configuration, giving precedence to user-defined values.

Constants are also defined for default filenames (e.g., DefaultConfigFileName,
DefaultCodemodFilename, DefaultDescribeOutputFile, UltimateDefaultAILogFilename)
and default patterns for validation. These constants are intended to be used by
CLI commands to ensure consistent default behavior.

The typical flow involves:
1. Attempting to find and load a user-defined '.contextvibes.yaml' file.
2. If found and valid, merging it with the application's default configuration.
3. If not found or invalid, using the application's default configuration directly.
The resulting configuration is then used throughout the application, particularly
by the cmd package to influence command behavior.
*/
package config

======== END FILE: internal/config/doc.go ========

======== FILE: internal/exec/client.go ========

// internal/exec/client.go
package exec

import (
	"context"
	"log/slog"
)

// ExecutorClient provides a high-level interface for running external commands.
// It uses an underlying CommandExecutor for the actual execution.
type ExecutorClient struct {
	executor CommandExecutor // unexported field
}

// NewClient creates a new ExecutorClient with the given CommandExecutor.
func NewClient(executor CommandExecutor) *ExecutorClient {
	return &ExecutorClient{executor: executor}
}

// Execute runs a command, typically piping stdio. See CommandExecutor.Execute.
func (c *ExecutorClient) Execute(ctx context.Context, dir string, commandName string, args ...string) error {
	return c.executor.Execute(ctx, dir, commandName, args...)
}

// CaptureOutput runs a command and captures its stdout and stderr. See CommandExecutor.CaptureOutput.
func (c *ExecutorClient) CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (string, string, error) {
	return c.executor.CaptureOutput(ctx, dir, commandName, args...)
}

// CommandExists checks if a command is available. See CommandExecutor.CommandExists.
func (c *ExecutorClient) CommandExists(commandName string) bool {
	return c.executor.CommandExists(commandName)
}

// Logger returns the logger from the underlying executor.
func (c *ExecutorClient) Logger() *slog.Logger {
	return c.executor.Logger()
}

// UnderlyingExecutor returns the CommandExecutor used by this client.
// This allows passing the raw executor to other components if needed.
func (c *ExecutorClient) UnderlyingExecutor() CommandExecutor { // New Exported Getter
    return c.executor
}

======== END FILE: internal/exec/client.go ========

======== FILE: internal/exec/doc.go ========

// internal/exec/doc.go

/*
Package exec provides a client and interface for executing external commands.
It abstracts the underlying os/exec calls, allowing for easier testing and
consistent command execution logic throughout the application.

The primary components are:
  - CommandExecutor: An interface defining methods to run commands and capture output.
  - OSCommandExecutor: The default implementation of CommandExecutor using os/exec.
  - ExecutorClient: A client that uses a CommandExecutor to provide higher-level
    methods for command execution.

Usage:

	// In your application setup (e.g., cmd/root.go or per command)
	osExecutor := exec.NewOSCommandExecutor(someLogger) // Pass an *slog.Logger
	execClient := exec.NewClient(osExecutor)

	// Later, to run a command:
	err := execClient.Execute(ctx, "/tmp", "ls", "-l")
	if err != nil {
		// handle error
	}

	// To capture output:
	stdout, stderr, err := execClient.CaptureOutput(ctx, ".", "go", "version")
	if err != nil {
		// handle error
	}
	fmt.Printf("Go version: %s", stdout)
*/
package exec

======== END FILE: internal/exec/doc.go ========

======== FILE: internal/exec/executor.go ========

// internal/exec/executor.go
package exec

import (
	"context"
	"log/slog"
)

// CommandExecutor defines the interface for running external commands.
// Implementations of this interface handle the actual execution logic.
type CommandExecutor interface {
	// Execute runs a command, connecting stdio to the parent process's stdio.
	// dir: the working directory for the command.
	// commandName: the name or path of the command to run.
	// args: arguments for the command.
	// Returns an error if execution fails.
	Execute(ctx context.Context, dir string, commandName string, args ...string) error

	// CaptureOutput runs a command, capturing its stdout and stderr.
	// dir: the working directory for the command.
	// commandName: the name or path of the command to run.
	// args: arguments for the command.
	// Returns stdout, stderr, and any error (including *exec.PkgExitError).
	CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (stdout, stderr string, err error)

	// CommandExists checks if a command is available in the PATH or at the specified path.
	CommandExists(commandName string) bool

	// Logger returns the logger associated with this executor.
	Logger() *slog.Logger
}

======== END FILE: internal/exec/executor.go ========

======== FILE: internal/exec/os_executor.go ========

// internal/exec/os_executor.go
package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec" // Standard library exec
	"strings"
)

// OSCommandExecutor is the default implementation of CommandExecutor using the os/exec package.
type OSCommandExecutor struct {
	logger *slog.Logger
}

// NewOSCommandExecutor creates a new OSCommandExecutor.
// If logger is nil, a discard logger will be used.
func NewOSCommandExecutor(logger *slog.Logger) CommandExecutor {
	log := logger
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil)) // Default to discard if no logger provided
	}
	return &OSCommandExecutor{logger: log}
}

func (e *OSCommandExecutor) Logger() *slog.Logger {
	return e.logger
}

func (e *OSCommandExecutor) Execute(ctx context.Context, dir string, commandName string, args ...string) error {
	e.logger.DebugContext(ctx, "Executing command",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args),
		slog.String("dir", dir))

	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout // Pipe directly
	cmd.Stderr = os.Stderr // Pipe directly
	cmd.Stdin = os.Stdin   // Pipe directly

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			e.logger.ErrorContext(ctx, "Command failed with exit code",
				slog.String("component", "OSCommandExecutor"),
				slog.String("command", commandName),
				slog.Any("args", args),
				slog.Int("exit_code", exitErr.ExitCode()),
				slog.String("error", err.Error()))
			// Stderr already piped. Return error that includes exit code info.
			return fmt.Errorf("command '%s %s' failed with exit code %d: %w", commandName, strings.Join(args, " "), exitErr.ExitCode(), err)
		}
		e.logger.ErrorContext(ctx, "Failed to execute command",
			slog.String("component", "OSCommandExecutor"),
			slog.String("command", commandName),
			slog.Any("args", args),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to start or execute command '%s %s': %w", commandName, strings.Join(args, " "), err)
	}

	e.logger.InfoContext(ctx, "Command executed successfully",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args))
	return nil
}

func (e *OSCommandExecutor) CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (string, string, error) {
	e.logger.DebugContext(ctx, "Capturing command output",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args),
		slog.String("dir", dir))

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, commandName, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if err != nil {
		var exitErr *exec.ExitError
		// Construct a more informative error message
		errMsg := fmt.Sprintf("command '%s %s' in dir '%s' failed", commandName, strings.Join(args, " "), dir)
		if errors.As(err, &exitErr) {
			errMsg = fmt.Sprintf("%s with exit code %d", errMsg, exitErr.ExitCode())
		} else {
			errMsg = fmt.Sprintf("%s: %v", errMsg, err) // Include original error type/msg for non-ExitErrors
		}

		trimmedStderr := strings.TrimSpace(stderrStr)
		if trimmedStderr != "" {
			errMsg = fmt.Sprintf("%s. Stderr: %s", errMsg, trimmedStderr)
		}

		e.logger.ErrorContext(ctx, "Command capture failed",
			slog.String("component", "OSCommandExecutor"),
			slog.String("command", commandName),
			slog.Any("args", args),
			slog.String("stdout_capture_len", fmt.Sprintf("%d bytes", len(stdoutStr))),
			slog.String("stderr_capture_len", fmt.Sprintf("%d bytes", len(stderrStr))),
			slog.String("error", err.Error()),     // Log the original simpler error
			slog.String("detailed_error", errMsg)) // Log the detailed constructed error
		return stdoutStr, stderrStr, fmt.Errorf(errMsg+": %w", err) // Wrap original error
	}

	e.logger.DebugContext(ctx, "Command capture successful",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName),
		slog.Any("args", args),
		slog.String("stdout_capture_len", fmt.Sprintf("%d bytes", len(stdoutStr))),
		slog.String("stderr_capture_len", fmt.Sprintf("%d bytes", len(stderrStr))))
	return stdoutStr, stderrStr, nil
}

func (e *OSCommandExecutor) CommandExists(commandName string) bool {
	_, err := exec.LookPath(commandName)
	if err != nil {
		e.logger.Debug("Command existence check: not found",
			slog.String("component", "OSCommandExecutor"),
			slog.String("command", commandName),
			slog.String("error", err.Error()))
		return false
	}
	e.logger.Debug("Command existence check: found",
		slog.String("component", "OSCommandExecutor"),
		slog.String("command", commandName))
	return true
}

======== END FILE: internal/exec/os_executor.go ========

======== FILE: internal/git/client.go ========

// internal/git/client.go
package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	osexec "os/exec" // Alias for standard library exec.ExitError
	"path/filepath"
	"strings"

	"github.com/contextvibes/cli/internal/exec" // Use the new executor
)

// GitClient provides methods for interacting with a Git repository.
type GitClient struct {
	repoPath string
	gitDir   string
	config   GitClientConfig
	logger   *slog.Logger         // Logger for GitClient's specific operations
	executor exec.CommandExecutor // Uses the new CommandExecutor from internal/exec
}

// NewClient creates and initializes a new GitClient.
// The GitClientConfig's Executor field should be pre-populated,
// or validateAndSetDefaults will create a default OSCommandExecutor.
func NewClient(ctx context.Context, workDir string, config GitClientConfig) (*GitClient, error) {
	// validateAndSetDefaults will ensure Logger and Executor are non-nil.
	validatedConfig, err := config.validateAndSetDefaults()
	if err != nil {
		return nil, fmt.Errorf("invalid GitClientConfig: %w", err)
	}
	logger := validatedConfig.Logger     // This is the GitClient's own logger
	executor := validatedConfig.Executor // This is the exec.CommandExecutor

	// Check if the configured git executable exists using the provided executor
	if !executor.CommandExists(validatedConfig.GitExecutable) {
		err := fmt.Errorf("git executable '%s' not found in PATH or specified path", validatedConfig.GitExecutable)
		logger.ErrorContext(ctx, "GitClient setup failed: executable check",
			slog.String("source_component", "GitClient.NewClient"),
			slog.String("error", err.Error()),
			slog.String("executable_path", validatedConfig.GitExecutable))
		return nil, err
	}

	effectiveWorkDir := workDir
	if effectiveWorkDir == "" {
		effectiveWorkDir, err = os.Getwd()
		if err != nil {
			logger.ErrorContext(ctx, "GitClient setup failed: getwd", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// Use the executor to find repo top-level
	topLevelCmdArgs := []string{"rev-parse", "--show-toplevel"}
	topLevel, stderr, err := executor.CaptureOutput(ctx, effectiveWorkDir, validatedConfig.GitExecutable, topLevelCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --show-toplevel",
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("initial_workdir", effectiveWorkDir))
		return nil, fmt.Errorf("path '%s' is not within a Git repository (or git command '%s' failed)", effectiveWorkDir, validatedConfig.GitExecutable)
	}
	repoPath := strings.TrimSpace(topLevel)

	// Use the executor to find .git directory
	gitDirCmdArgs := []string{"rev-parse", "--git-dir"}
	gitDirOutput, stderr, err := executor.CaptureOutput(ctx, repoPath, validatedConfig.GitExecutable, gitDirCmdArgs...)
	if err != nil {
		logger.ErrorContext(ctx, "GitClient setup failed: rev-parse --git-dir",
			slog.String("error", err.Error()),
			slog.String("stderr", strings.TrimSpace(stderr)),
			slog.String("repo_path", repoPath))
		return nil, fmt.Errorf("could not determine .git directory for repo at '%s': %w", repoPath, err)
	}
	gitDir := strings.TrimSpace(gitDirOutput)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoPath, gitDir)
	}

	client := &GitClient{
		repoPath: repoPath,
		gitDir:   gitDir,
		config:   validatedConfig,
		logger:   logger,
		executor: executor,
	}
	logger.InfoContext(ctx, "GitClient initialized successfully",
		slog.String("repository_path", client.repoPath),
		slog.String("git_dir", client.gitDir))
	return client, nil
}

func (c *GitClient) Path() string           { return c.repoPath }
func (c *GitClient) GitDir() string         { return c.gitDir }
func (c *GitClient) MainBranchName() string { return c.config.DefaultMainBranchName }
func (c *GitClient) RemoteName() string     { return c.config.DefaultRemoteName }
func (c *GitClient) Logger() *slog.Logger   { return c.logger }

func (c *GitClient) runGit(ctx context.Context, args ...string) error {
	// Logger().Debug(...) is already part of executor.Execute
	return c.executor.Execute(ctx, c.repoPath, c.config.GitExecutable, args...)
}

func (c *GitClient) captureGitOutput(ctx context.Context, args ...string) (string, string, error) {
	// Logger().Debug(...) is already part of executor.CaptureOutput
	return c.executor.CaptureOutput(ctx, c.repoPath, c.config.GitExecutable, args...)
}

// --- Public Git Operation Methods ---
// (These methods remain largely the same but now internally call c.executor methods
//  which are of type exec.CommandExecutor)

func (c *GitClient) GetCurrentBranchName(ctx context.Context) (string, error) {
	stdout, stderr, err := c.captureGitOutput(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to get current branch name", "error", err, "stderr", strings.TrimSpace(stderr))
		return "", fmt.Errorf("failed to get current branch name: %w", err)
	}
	branch := strings.TrimSpace(stdout)
	if branch == "HEAD" {
		return "", fmt.Errorf("currently in detached HEAD state")
	}
	if branch == "" {
		return "", fmt.Errorf("could not determine current branch name (empty output)")
	}
	return branch, nil
}

func (c *GitClient) AddAll(ctx context.Context) error {
	err := c.runGit(ctx, "add", ".")
	if err != nil {
		return fmt.Errorf("git add . failed: %w", err)
	}
	return nil
}

func (c *GitClient) Commit(ctx context.Context, message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}
	err := c.runGit(ctx, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("commit command failed: %w", err)
	}
	return nil
}

func (c *GitClient) HasStagedChanges(ctx context.Context) (bool, error) {
	_, _, err := c.captureGitOutput(ctx, "diff", "--quiet", "--cached")
	if err == nil {
		return false, nil // Exit 0: no changes
	}
	var exitErr *osexec.ExitError // Use osexec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil // Exit 1: changes found
	}
	return false, fmt.Errorf("failed to determine staged status: %w", err)
}

func (c *GitClient) GetStatusShort(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "status", "--short")
}

func (c *GitClient) GetDiffCached(ctx context.Context) (string, string, error) {
	stdout, stderr, err := c.captureGitOutput(ctx, "diff", "--cached")
	if err != nil {
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return stdout, stderr, nil // Exit 1 (changes found) is not an error for this func
		}
		return stdout, stderr, err // Actual error
	}
	return stdout, stderr, nil // No error, no diff
}

func (c *GitClient) GetDiffUnstaged(ctx context.Context) (string, string, error) {
	stdout, stderr, err := c.captureGitOutput(ctx, "diff", "HEAD")
	if err != nil {
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return stdout, stderr, nil // Exit 1 (changes found) is not an error for this func
		}
		return stdout, stderr, err // Actual error
	}
	return stdout, stderr, nil // No error, no diff
}

func (c *GitClient) ListUntrackedFiles(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "ls-files", "--others", "--exclude-standard")
}

func (c *GitClient) IsWorkingDirClean(ctx context.Context) (bool, error) {
	_, _, errDiff := c.captureGitOutput(ctx, "diff", "--quiet")
	if errDiff != nil {
		var exitErr *osexec.ExitError
		if errors.As(errDiff, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil // Unstaged changes
		}
		return false, fmt.Errorf("failed checking unstaged changes: %w", errDiff)
	}
	hasStaged, errStaged := c.HasStagedChanges(ctx)
	if errStaged != nil {
		return false, fmt.Errorf("failed checking staged changes: %w", errStaged)
	}
	if hasStaged {
		return false, nil // Staged changes
	}
	untrackedOut, _, errUntracked := c.ListUntrackedFiles(ctx)
	if errUntracked != nil {
		return false, fmt.Errorf("failed checking untracked files: %w", errUntracked)
	}
	if strings.TrimSpace(untrackedOut) != "" {
		return false, nil // Untracked files
	}
	return true, nil
}

func (c *GitClient) PullRebase(ctx context.Context, branch string) error {
	remote := c.RemoteName()
	err := c.runGit(ctx, "pull", "--rebase", remote, branch)
	if err != nil {
		return fmt.Errorf("git pull --rebase %s %s failed: %w", remote, branch, err)
	}
	return nil
}

func (c *GitClient) IsBranchAhead(ctx context.Context) (bool, error) {
	stdout, _, err := c.captureGitOutput(ctx, "status", "-sb")
	if err != nil {
		return false, fmt.Errorf("failed to get status to check if branch is ahead: %w", err)
	}
	return strings.Contains(stdout, "[ahead "), nil
}

func (c *GitClient) Push(ctx context.Context, branch string) error {
	remote := c.RemoteName()
	args := []string{"push", remote}
	if branch != "" {
		args = append(args, branch)
	}
	// Capture output to check for "up-to-date" messages, as runGit only returns error on non-zero exit.
	_, stderr, err := c.captureGitOutput(ctx, args...) // Use captureGitOutput
	if err != nil {
		// Check if stderr (or err.Error() if it includes stderr) indicates "up-to-date"
		// This is a bit fragile. A better way would be for CaptureOutput to return specific error types.
		errMsg := strings.ToLower(err.Error() + " " + stderr) // Combine for checking
		if strings.Contains(errMsg, "everything up-to-date") || strings.Contains(errMsg, "already up-to-date") {
			c.logger.InfoContext(ctx, "'git push' reported everything up-to-date.", "remote", remote, "branch_arg", branch)
			return nil // Not a failure
		}
		return fmt.Errorf("git push failed: %w", err)
	}
	return nil
}

func (c *GitClient) LocalBranchExists(ctx context.Context, branchName string) (bool, error) {
	ref := fmt.Sprintf("refs/heads/%s", branchName)
	_, _, err := c.captureGitOutput(ctx, "show-ref", "--verify", "--quiet", ref)
	if err == nil {
		return true, nil // Exit 0 means it exists
	}
	var exitErr *osexec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil // Exit 1 means it doesn't exist
	}
	return false, fmt.Errorf("failed to check existence of local branch '%s': %w", branchName, err)
}

func (c *GitClient) SwitchBranch(ctx context.Context, branchName string) error {
	err := c.runGit(ctx, "switch", branchName)
	if err != nil {
		return fmt.Errorf("git switch %s failed: %w", branchName, err)
	}
	return nil
}

func (c *GitClient) CreateAndSwitchBranch(ctx context.Context, newBranchName string, baseBranch string) error {
	args := []string{"switch", "-c", newBranchName}
	if baseBranch != "" {
		args = append(args, baseBranch)
	}
	err := c.runGit(ctx, args...)
	if err != nil {
		return fmt.Errorf("git switch -c %s failed: %w", newBranchName, err)
	}
	return nil
}

func (c *GitClient) PushAndSetUpstream(ctx context.Context, branchName string) error {
	remote := c.RemoteName()
	err := c.runGit(ctx, "push", "--set-upstream", remote, branchName)
	if err != nil {
		return fmt.Errorf("git push --set-upstream %s %s failed: %w", remote, branchName, err)
	}
	return nil
}

func (c *GitClient) ListTrackedAndCachedFiles(ctx context.Context) (string, string, error) {
	return c.captureGitOutput(ctx, "ls-files", "-co", "--exclude-standard")
}

// TruncateString helper can be removed if not used, or kept if useful elsewhere.
// For now, keeping it as it was in the provided file.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		if maxLen < 0 {
			return ""
		}
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

======== END FILE: internal/git/client.go ========

======== FILE: internal/git/config.go ========

// internal/git/config.go
package git

import (
	// "io" // No longer needed for discard logger if logger is required
	"log/slog"

	"github.com/contextvibes/cli/internal/exec" // Import the new exec package
)

type GitClientConfig struct {
	GitExecutable         string
	DefaultRemoteName     string               // Will be set by cmd layer from LoadedAppConfig
	DefaultMainBranchName string               // Will be set by cmd layer from LoadedAppConfig
	Logger                *slog.Logger         // Logger for GitClient's own operations
	Executor              exec.CommandExecutor // Use the new CommandExecutor interface from internal/exec
}

// validateAndSetDefaults now expects that if an Executor is needed, it's either provided,
// or can be created using a logger that should also be provided (or defaulted).
// The primary logger for the application (AppLogger) can be passed to create a default executor.
func (c GitClientConfig) validateAndSetDefaults() (GitClientConfig, error) {
	validated := c

	// DefaultRemoteName and DefaultMainBranchName are expected to be set by the caller (cmd/root.go)
	// using the application's LoadedAppConfig. We still provide fallbacks here for safety.
	if validated.DefaultRemoteName == "" {
		validated.DefaultRemoteName = "origin"
	}
	if validated.DefaultMainBranchName == "" {
		validated.DefaultMainBranchName = "main"
	}
	if validated.GitExecutable == "" {
		validated.GitExecutable = "git" // Default to looking for 'git' in PATH
	}

	// Logger for the GitClient itself. If not provided, it might inherit from the Executor's logger,
	// or use a specific one. For simplicity, let's ensure it has one.
	// The AppLogger from cmd/root.go is a good candidate to pass into here.
	if validated.Logger == nil {
		// This state (nil Logger in GitClientConfig) should ideally be avoided by the caller.
		// If the Executor is also nil, its creation below would also lack a logger.
		// For robustness, if Executor is also nil, its default creation will use its own discard/default logger.
		// If Executor is provided, GitClient can use its logger.
		if validated.Executor != nil {
			validated.Logger = validated.Executor.Logger()
		} else {
			// Fallback: create a new OS executor which will have its own default/discard logger
			// and use that logger for the GitClient. This is less ideal than injecting AppLogger.
			tempExecutor := exec.NewOSCommandExecutor(nil) // Creates OS executor with a discard logger
			validated.Logger = tempExecutor.Logger()
		}
	}

	// Executor for running git commands.
	if validated.Executor == nil {
		// If no executor is provided, create a default OSCommandExecutor.
		// It's crucial that this OSCommandExecutor uses a proper logger.
		// Pass the GitClient's logger (which should have been resolved above, possibly from AppLogger).
		validated.Executor = exec.NewOSCommandExecutor(validated.Logger)
	}

	return validated, nil
}

======== END FILE: internal/git/config.go ========

======== FILE: internal/git/doc.go ========

// internal/git/doc.go

/*
Package git provides a high-level client for interacting with Git repositories
programmatically from Go applications. It abstracts the direct execution of 'git'
command-line operations, offering a more Go-idiomatic API.

The primary entry point for using this package is the GitClient type, which
is instantiated via the NewClient function. The client requires a working directory
to determine the repository context and can be configured using GitClientConfig.

Key features include:
  - Repository context detection (finding .git and top-level directories).
  - Execution of common Git commands (status, commit, add, branch, etc.)
    through structured methods.
  - Abstraction over command execution, allowing for custom executors (primarily for testing).
  - Integration with structured logging via the slog package.

Usage Example:

	// Assume globalLogger is an initialized *slog.Logger
	ctx := context.Background()
	workDir, _ := os.Getwd()

	gitCfg := git.GitClientConfig{
		Logger: globalLogger,
		// Other configurations can be set here
	}

	client, err := git.NewClient(ctx, workDir, gitCfg)
	if err != nil {
		log.Fatalf("Failed to create Git client: %v", err)
	}

	branch, err := client.GetCurrentBranchName(ctx)
	if err != nil {
		log.Printf("Error getting branch: %v", err)
	} else {
		log.Printf("Current branch: %s", branch)
	}

Error Handling:

Methods on the GitClient typically return an error as their last argument if an
operation fails. Errors can originate from underlying git command failures,
invalid input, or issues with the repository state. It's important for callers
to check these errors.

Logging:

The GitClient uses an slog.Logger instance provided via GitClientConfig.
This allows for consistent, structured logging of its operations, which can be
directed to various outputs (e.g., console, files, AI-consumable streams)
by configuring the logger's handlers at the application level.

Testing:

The GitClient is designed with testability in mind. The GitClientConfig.Executor
field allows injecting a mock 'executor' interface, enabling unit tests for
client methods without relying on an actual 'git' executable or a live repository.
*/
package git

======== END FILE: internal/git/doc.go ========

======== FILE: internal/project/detector.go ========

package project

import (
	"fmt"
	"os"
	"path/filepath"
)

func Detect(dir string) (Type, error) {
	tfFiles, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return Unknown, fmt.Errorf("error checking for Terraform files: %w", err)
	}
	if len(tfFiles) > 0 {
		return Terraform, nil
	}

	pulumiYamlPath := filepath.Join(dir, "Pulumi.yaml")
	if _, err := os.Stat(pulumiYamlPath); err == nil {
		return Pulumi, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for Pulumi.yaml: %w", err)
	}

	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		return Go, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for go.mod: %w", err)
	}

	pyReqPath := filepath.Join(dir, "requirements.txt")
	pyProjPath := filepath.Join(dir, "pyproject.toml")
	if _, err := os.Stat(pyReqPath); err == nil {
		return Python, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for requirements.txt: %w", err)
	}
	if _, err := os.Stat(pyProjPath); err == nil {
		return Python, nil
	} else if !os.IsNotExist(err) {
		return Unknown, fmt.Errorf("error checking for pyproject.toml: %w", err)
	}

	return Unknown, nil
}

======== END FILE: internal/project/detector.go ========

======== FILE: internal/project/types.go ========

package project

// Type represents the detected project type.
type Type string

const (
	Terraform Type = "Terraform"
	Pulumi    Type = "Pulumi"
	Go        Type = "Go"
	Python    Type = "Python"
	Unknown   Type = "Unknown"
)

======== END FILE: internal/project/types.go ========

======== FILE: internal/tools/io.go ========

package tools

import (
	"bytes"
	"fmt"
	"os"
)

// ReadFileContent reads the entire content of the file at the specified path.
// Returns the content as a byte slice or an error if reading fails.
func ReadFileContent(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Wrap the error with more context.
		return nil, fmt.Errorf("error reading file '%s': %w", filePath, err)
	}
	return content, nil
}

// WriteBufferToFile writes the content of a bytes.Buffer to the specified file path.
// It uses default file permissions (0644).
// It prints informational messages about writing to os.Stdout.
// TODO: Refactor to remove direct fmt.Printf calls.
//
//	Calling commands should use their Presenter for user-facing messages
//	or a Logger for debug/trace information related to file writing.
//	This function should focus solely on writing the file.
func WriteBufferToFile(filePath string, buf *bytes.Buffer) error {
	// These fmt.Printf calls directly write to os.Stdout.
	// They are currently used by cmd/diff and cmd/describe.
	// Future refactoring might replace these with presenter calls from the cmd layer.
	fmt.Printf("INFO: Writing output to %s...\n", filePath)
	err := os.WriteFile(filePath, buf.Bytes(), 0644) // Use standard file permissions
	if err != nil {
		// Wrap the error with more context.
		return fmt.Errorf("failed to write output file '%s': %w", filePath, err)
	}
	fmt.Printf("INFO: Successfully wrote %s.\n", filePath)
	return nil
}

======== END FILE: internal/tools/io.go ========

======== FILE: internal/tools/markdown.go ========

package tools

import (
	"bytes"
	"fmt"
	"strings"
)

// AppendSectionHeader adds a standard Markdown H3 section header to the buffer.
func AppendSectionHeader(buf *bytes.Buffer, title string) {
	buf.WriteString("### ")
	buf.WriteString(title)
	buf.WriteString("\n\n")
}

// AppendFencedCodeBlock adds a standard Markdown fenced code block to the buffer.
func AppendFencedCodeBlock(buf *bytes.Buffer, content string, languageHint string) {
	buf.WriteString("```")
	if languageHint != "" {
		buf.WriteString(languageHint)
	}
	buf.WriteString("\n")
	// Ensure content ends with a newline before the closing fence
	// But avoid adding a double newline if one already exists
	trimmedContent := strings.TrimSuffix(content, "\n")
	buf.WriteString(trimmedContent)
	buf.WriteByte('\n') // Ensure at least one newline

	buf.WriteString("```\n\n")
}

// AppendFileMarkerHeader adds the explicit file start marker.
func AppendFileMarkerHeader(buf *bytes.Buffer, filePath string) {
	// Ensure preceding content has adequate spacing, but avoid excessive newlines
	trimmedBytes := bytes.TrimRight(buf.Bytes(), "\n")
	buf.Reset()
	buf.Write(trimmedBytes)
	// Add consistent spacing before the header
	fmt.Fprintf(buf, "\n\n======== FILE: %s ========\n\n", filePath)
}

// AppendFileMarkerFooter adds the explicit file end marker.
func AppendFileMarkerFooter(buf *bytes.Buffer, filePath string) {
	// Simpler approach: Trim all trailing whitespace, then add exactly two newlines before the footer.
	trimmedBytes := bytes.TrimSpace(buf.Bytes())
	buf.Reset()             // Clear the buffer
	buf.Write(trimmedBytes) // Write back the trimmed content
	// Add exactly two newlines before the footer marker
	fmt.Fprintf(buf, "\n\n======== END FILE: %s ========\n\n", filePath)
}

======== END FILE: internal/tools/markdown.go ========

======== FILE: internal/ui/presenter.go ========

// internal/ui/presenter.go

package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color" // Import the color library
)

// Presenter handles structured writing to standard output and standard error,
// and reading standardized user input, mimicking the Pulumi CLI style.
type Presenter struct {
	// Interface fields for flexibility and testing
	outW io.Writer
	errW io.Writer
	inR  io.Reader

	// Color instances (initialized in New)
	successColor *color.Color
	errorColor   *color.Color
	warningColor *color.Color
	infoColor    *color.Color
	stepColor    *color.Color
	detailColor  *color.Color
	promptColor  *color.Color
	headerColor  *color.Color
	boldColor    *color.Color
	summaryColor *color.Color
}

// NewPresenter creates a new Console instance with Pulumi-like color support.
// Color support is automatically detected and disabled if the terminal doesn't support it
// or if the NO_COLOR environment variable is set.
// If outW, errW, or inR are nil, they default to os.Stdout, os.Stderr, and os.Stdin respectively.
func NewPresenter(outW, errW io.Writer, inR io.Reader) *Presenter {
	// *** CORRECTED VARIABLE DECLARATIONS AND ASSIGNMENTS ***
	// Declare local variables with the correct INTERFACE types
	var out io.Writer = os.Stdout
	var err io.Writer = os.Stderr
	var in io.Reader = os.Stdin

	// Assign parameters ONLY if they are not nil, overwriting defaults
	if outW != nil {
		out = outW
	}
	if errW != nil {
		err = errW
	}
	if inR != nil {
		in = inR
	}
	// *********************************************************

	// Initialize and return the struct, assigning interface values to interface fields
	return &Presenter{
		outW: out, // Assign io.Writer to io.Writer field
		errW: err, // Assign io.Writer to io.Writer field
		inR:  in,  // Assign io.Reader to io.Reader field

		// Initialize all color fields
		successColor: color.New(color.FgGreen, color.Bold),
		errorColor:   color.New(color.FgRed, color.Bold),
		warningColor: color.New(color.FgYellow),
		infoColor:    color.New(color.FgBlue),
		stepColor:    color.New(color.FgWhite),
		detailColor:  color.New(color.Faint),
		promptColor:  color.New(color.FgCyan),
		headerColor:  color.New(color.Bold, color.Underline),
		boldColor:    color.New(color.Bold),
		summaryColor: color.New(color.FgCyan, color.Bold),
	}
}

// --- Output Stream Getters ---

// Out returns the configured output writer (typically os.Stdout).
func (p *Presenter) Out() io.Writer {
	return p.outW
}

// Err returns the configured error writer (typically os.Stderr).
func (p *Presenter) Err() io.Writer {
	return p.errW
}

// --- Output Formatting Methods ---
// These methods correctly use p.outW and p.errW which are io.Writer interfaces

func (p *Presenter) Header(format string, a ...any) { p.headerColor.Fprintf(p.outW, format+"\n", a...) }
func (p *Presenter) Summary(format string, a ...any) {
	p.summaryColor.Fprint(p.outW, "SUMMARY:\n")
	fmt.Fprintf(p.outW, "  "+format+"\n", a...)
	p.Newline()
}
func (p *Presenter) Step(format string, a ...any) {
	p.stepColor.Fprintf(p.outW, "- "+format+"\n", a...)
}
func (p *Presenter) Info(format string, a ...any) {
	p.infoColor.Fprintf(p.outW, "~ "+format+"\n", a...)
}
func (p *Presenter) InfoPrefixOnly() { p.infoColor.Fprint(p.outW, "~ ") }
func (p *Presenter) Success(format string, a ...any) {
	p.successColor.Fprintf(p.outW, "+ "+format+"\n", a...)
}
func (p *Presenter) Error(format string, a ...any) {
	p.errorColor.Fprintf(p.errW, "! "+format+"\n", a...)
}
func (p *Presenter) Warning(format string, a ...any) {
	p.warningColor.Fprintf(p.errW, "~ "+format+"\n", a...)
}
func (p *Presenter) Advice(format string, a ...any) {
	p.warningColor.Fprintf(p.outW, "~ "+format+"\n", a...)
}
func (p *Presenter) Detail(format string, a ...any) {
	p.detailColor.Fprintf(p.outW, "  "+format+"\n", a...)
}
func (p *Presenter) Highlight(text string) string { return p.boldColor.Sprint(text) }
func (p *Presenter) Newline()                     { fmt.Fprintln(p.outW) }
func (p *Presenter) Separator() {
	color.New(color.Faint).Fprintln(p.outW, "----------------------------------------")
}

// --- Input Methods ---
// These methods correctly use p.inR which is an io.Reader interface

func (p *Presenter) PromptForInput(prompt string) (string, error) {
	reader := bufio.NewReader(p.inR) // Use interface field
	prompt = strings.TrimSpace(prompt)
	if !strings.HasSuffix(prompt, ":") {
		prompt += ":"
	}
	prompt += " "
	p.promptColor.Fprint(p.errW, prompt) // Write prompt to error stream
	input, err := reader.ReadString('\n')
	if err != nil {
		p.errorColor.Fprintf(p.errW, "\n! Error reading input: %v\n", err)
		return "", fmt.Errorf("reading input failed: %w", err)
	}
	return strings.TrimSpace(input), nil
}

func (p *Presenter) PromptForConfirmation(prompt string) (bool, error) {
	reader := bufio.NewReader(p.inR) // Use interface field
	prompt = strings.TrimSpace(prompt)
	if !strings.HasSuffix(prompt, "?") {
		prompt += "?"
	}
	fullPrompt := prompt + " [y/N]: "
	for {
		p.promptColor.Fprint(p.errW, fullPrompt) // Write prompt to error stream
		input, err := reader.ReadString('\n')
		if err != nil {
			p.errorColor.Fprintf(p.errW, "\n! Error reading confirmation: %v\n", err)
			return false, fmt.Errorf("reading confirmation failed: %w", err)
		}
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			return true, nil
		}
		if input == "n" || input == "no" || input == "" {
			return false, nil
		}
		p.warningColor.Fprintf(p.errW, "~ Invalid input. Please enter 'y' or 'n'.\n") // Write warning to error stream
	}
}

======== END FILE: internal/ui/presenter.go ========

