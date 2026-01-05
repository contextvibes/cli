# Changelog

All notable changes to the **Context Vibes CLI** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Deprecated
- **\`craft\` Command Pillar:** The \`craft\` command group has been deprecated. Its functionality has been integrated directly into the core \`factory\` and \`product\` commands to streamline the workflow.

### Added
- **AI-Assisted Flags (`--ai`):** Added a global-style \`--ai\` flag to key commands to generate optimized prompts for external AI assistants:
  - \`factory commit --ai\`: Replaces \`craft message\`. Generates a Conventional Commit message prompt based on staged changes.
  - \`factory finish --ai\`: Replaces \`craft pr-description\`. Generates a Pull Request description prompt based on branch diffs.
  - \`product quality --ai\`: Replaces \`craft review\`. Generates a Code Review prompt for the codebase.
  - \`product codemod --ai\`: Replaces \`craft refactor\`. Generates a Refactoring Plan prompt for specific files.
- **Strategic Kickoff Integration:**
  - \`factory kickoff --strategic\`: Replaces \`craft kickoff\`. Generates the master protocol for an AI-guided strategic project kickoff.

### Changed
- **Workflow Architecture:** Refactored internal workflow steps to be reusable across different commands, enabling the "AI as a feature" model.

---

## [0.6.0] - 2025-12-05

### Added
- **New \`factory tools\` Command:**
  - Force-rebuilds development tools (\`govulncheck\`, \`golangci-lint\`) to match the system Go version.
  - Automatically configures \`.bashrc\` to prioritize local tools.
  - Resolves version mismatches often seen in Nix environments.

## [0.4.1] - 2025-12-02

### Changed
- **Architecture Refactor:**
  - Introduced \`internal/pipeline\` package for modular, testable quality checks.
  - Introduced \`internal/workflow\` package for reusable, step-based CLI workflows.
  - Refactored \`cmd/product/quality\` to use the new pipeline engine.
  - Refactored \`cmd/craft\` commands to use the new workflow engine.
- **Code Quality & Hygiene:**
  - Removed dead code and unused packages (\`internal/bootstrap\`, \`internal/contextgen\`).
  - Addressed \`noinlineerr\` and other linter violations.
  - \`quality\` command now automatically removes the stale \`_contextvibes.md\` report if all checks pass.

## [0.4.0] - 2025-12-01

### Added
- **New \`factory scrub\` Command:**
  - A "Scorched Earth" cleanup tool for the development environment.
  - Cleans Android artifacts, Go caches, Docker system (prune), Nix garbage, and user caches.
  - Includes safety prompts to prevent accidental data loss.
- **Enhanced \`product format\` Command:**
  - Now accepts specific file paths or directories as arguments (e.g., \`contextvibes product format cmd/root.go\`), allowing for faster feedback loops.
- **Environment Modernization:**
  - Upgraded project foundation to **Go 1.25**.
  - Updated Nix channel to \`stable-25.11\`.
  - Added \`gcc\` to the environment to support CGO tools (Delve, gopls).

### Changed
- **Code Quality Overhaul:**
  - Addressed over 400 linter violations.
  - Hardened \`.golangci.yml\` with stricter rules (magic numbers, error wrapping, global variables).
  - Fixed critical bugs in \`deploy\`, \`index\`, and config loading logic (handling of nil errors).
- **Configuration:**
  - Removed unnecessary \`replace\` directive from \`go.mod\`.

## [0.3.0] - 2025-08-11

### Added
- **New \`init\` Command:**
  - \`contextvibes init\` now creates a default \`.contextvibes.yaml\` file in the project root if one doesn't exist, streamlining project setup.
- **New \`project\` Command:**
  - \`contextvibes project list-issues\` fetches and displays all open issues, including comments, from the current GitHub repository.
  - Added an \`--output\` (\`-o\`) flag to save the formatted issue list to a file.
- **New Configuration Option for \`export\`:**
  - Added \`export.excludePatterns\` to \`.contextvibes.yaml\` to allow exclusion of files/directories (e.g., \`vendor/\`) from the \`context export all\` command.

### Changed
- **Refactored \`kickoff\` Command:**
  - Now depends on an interface for its presenter, significantly improving testability.
- **Improved \`export\` Command:**
  - Logic updated to respect the new \`export.excludePatterns\` configuration.
  - Now correctly detects and skips binary files to prevent corrupting the \`context_export_project.md\` output file.
- **Improved \`build\` Command:**
  - Refactored to respect the command's output streams, making it fully testable.

### Fixed
- **CLI Help Output:** Corrected a bug in \`cmd/root.go\` that caused command descriptions to be duplicated in the \`--help\` output.
- **\`test\` Command Flag Parsing:** Fixed a bug where flags like \`-v\` were not correctly passed to the underlying test runner (e.g., \`go test\`).
- **Unit Test Suite:**
  - Fixed multiple failing unit tests in \`cmd/build_test.go\` and \`internal/kickoff/orchestrator_test.go\`.
  - Resolved numerous \`golangci-lint\` warnings, including \`errcheck\`, \`unused\`, and \`stdmethods\`.

---
## [0.2.0] - 2025-06-08

### Added
- **\`contextvibes index\` Command Enhancement:**
    - Now generates a comprehensive JSON manifest (\`project_manifest.json\` by default) from local Markdown files.
    - Parses rich YAML front matter including \`title\`, \`artifactVersion\`, \`summary\`, \`usageGuidance\`, \`owner\`, \`createdDate\`, \`lastModifiedDate\`, \`defaultTargetPath\`, and \`tags\`.
    - Derives \`id\` (from relative path) and \`fileExtension\`.
    - \`lastModifiedDate\` falls back to file system modification time if not in front matter.
    - \`defaultTargetPath\` defaults to \`id.fileExtension\` if not in front matter.
- **New Command: \`contextvibes thea get-artifact <artifact-id>\`:**
    - Fetches a specific artifact document (e.g., playbook, template) from the central THEA framework repository.
    - Uses a hardcoded default URL for the THEA manifest (\`https://raw.githubusercontent.com/contextvibes/THEA/main/thea-manifest.json\`) and artifacts.
    - Supports \`--version\` flag for version hints (maps to Git ref).
    - Supports \`--output\` or \`-o\` to specify the local save path. Defaults to a name derived from artifact metadata if \`-o\` is omitted.
    - Supports \`--force\` or \`-f\` to overwrite existing output files.
- **Internal \`thea.Client\`:**
    - Added a new internal client (\`internal/thea/client.go\`) responsible for interacting with the THEA framework repository.
    - Capable of fetching and parsing the remote \`thea-manifest.json\`.
    - Capable of fetching specific artifact content based on manifest details.
    - Includes integration tests against the live THEA repository.

### Changed
- (List any significant changes to existing functionality if applicable)
- Example: If \`kickoff\` was refactored to use the \`thea.Client\` in this release, mention it here. (Sounds like we deferred this part for now).

### Fixed
- (List any bug fixes, e.g., "Resolved \`EOF\` errors in \`kickoff\` command unit tests by mocking Stdin for collaboration preference prompts." - *IF you fixed these*)

### Removed
- (List any features removed)

## [0.1.1] - 2025-05-10

### Added
- **Enhanced \`kickoff\` Command (Dual Mode Functionality):**
    - **Strategic Kickoff Prompt Generation Mode** (via \`contextvibes kickoff --strategic\` or on first run in a new project):
        - Conducts a brief interactive session to gather initial project details (name, type, stage) and user preferences for CLI interaction styles (\`ai.collaborationPreferences\`).
        - Generates a comprehensive master prompt file (\`STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md\`) by embedding and parameterizing a detailed protocol template (from \`internal/kickoff/assets/strategic_kickoff_protocol_template.md\`).
        - The generated master prompt is designed for the user to take to an external AI assistant (e.g., Gemini, Claude). It instructs the AI to guide the user through a strategic kickoff checklist, request further context via \`contextvibes\` commands, and generate structured YAML for specified configurations.
        - Saves gathered \`ai.collaborationPreferences\` to \`.contextvibes.yaml\` after the initial setup questions (before master prompt generation and also when marking complete).
    - **Mechanism to Mark Strategic Kickoff as Complete:**
        - New \`contextvibes kickoff --mark-strategic-complete\` flag.
        - This command updates \`.contextvibes.yaml\` by setting \`projectState.strategicKickoffCompleted: true\` and \`projectState.lastStrategicKickoffDate\`. It also ensures any \`ai.collaborationPreferences\` gathered during a preceding \`--strategic\` run's setup phase are persisted.
    - **Daily Git Workflow Mode:**
        - Becomes the default for \`contextvibes kickoff\` (without \`--strategic\`) once a strategic kickoff has been marked complete for the project.
        - Includes logic for prerequisite checks (clean working directory, on main branch), prompting for/validating branch name (respecting \`.contextvibes.yaml\`), and executing Git operations (pull rebase, new branch, push upstream).
        - Respects the global \`--yes\` flag for non-interactive operation.
- **New Configuration Options in \`.contextvibes.yaml\`:**
    - \`projectState\` section:
        - \`strategicKickoffCompleted\` (boolean, default: \`false\`)
        - \`lastStrategicKickoffDate\` (string, RFC3339 timestamp)
    - \`ai.collaborationPreferences\` section:
        - \`codeProvisioningStyle\` (string, default: "bash_cat_eof")
        - \`markdownDocsStyle\` (string, default: "raw_markdown")
        - \`detailedTaskMode\` (string, default: "mode_b")
        - \`proactiveDetailLevel\` (string, default: "detailed_explanations" or "concise_unless_asked" based on taskMode)
        - \`aiProactivity\` (string, default: "proactive_suggestions")
- **New Internal Package \`internal/kickoff\`:**
    - Contains \`Orchestrator\` to manage all \`kickoff\` command logic (strategic and daily).
    - Includes embedded master kickoff protocol template (\`internal/kickoff/assets/strategic_kickoff_protocol_template.md\`).
- **New Documentation:**
    - \`docs/PROJECT_KICKOFF_GUIDE.md\` explaining the new strategic kickoff workflow.
    - Initial \`DEVELOPMENT.md\` and \`CONTRIBUTING.md\`.
- **Sample \`.contextvibes.yaml\`:** Added to the project root as an example.

### Changed
- **\`cmd/kickoff.go\`:** Major refactor to use the new \`internal/kickoff.Orchestrator\` and handle new flags (\`--strategic\`, \`--mark-strategic-complete\`).
- **\`internal/config/config.go\`:**
    - \`Config\` struct extended with \`ProjectState\` and \`AISettings\`.
    - \`GetDefaultConfig()\` updated with defaults for new AI collaboration preferences and project state.
    - \`MergeWithDefaults()\` updated to correctly handle merging of new nested structs (field-by-field for \`AICollaborationPreferences\`).
    - \`UpdateAndSaveConfig()\` now uses an atomic write pattern (write to temp file then rename) for improved robustness.
- **\`internal/kickoff/orchestrator.go\`:** \`assumeYes\` flag handling refined to be a field within the \`Orchestrator\` struct, initialized from the global flag.

### Fixed
- Various minor compilation and linter issues encountered during the development of the enhanced kickoff feature.
- Ensured \`assumeYes\` is correctly plumbed from \`cmd/root.go\` to \`internal/kickoff.Orchestrator\` and respected in daily Git workflows.

---

## [0.0.5] - 2025-05-07

### Changed
*   **\`commit\` command:** Now fully supports configurable commit message validation via \`.contextvibes.yaml\`.
    *   Respects \`validation.commitMessage.enable\` to toggle validation.
    *   Uses \`validation.commitMessage.pattern\` for custom regex, falling back to the default Conventional Commits pattern if enabled and no custom pattern is provided.
    *   User feedback and help text updated to reflect active validation rules.
*   **\`cmd/root.go\`:** Improved the main help text (\`Long\` description) for the CLI to be more descriptive and structured.
*   Internal: Default AI log filename constant in \`internal/config/config.go\` (UltimateDefaultAILogFilename) aligned to \`contextvibes_ai_trace.log\` to match README documentation.
*   Internal: Default codemod script filename constant in \`internal/config/config.go\` (DefaultCodemodFilename) aligned to \`codemod.json\` for consistency.

### Fixed
*   **Commit Message Validation Regex:** The default commit message validation pattern (\`DefaultCommitMessagePattern\` in \`internal/config/config.go\`) now correctly allows \`/\` characters within the scope (e.g., \`feat(cmd/commit): ...\`), ensuring compatibility with common scope naming conventions.

---

## [0.0.4] - 2025-05-07

### Added

*   New \`codemod\` command (\`contextvibes codemod\`) to apply structured changes from a JSON script.
    *   Initially supports \`regex_replace\` and \`delete_file\` operations.
    *   Looks for \`codemod.json\` by default if \`--script\` flag is not provided.
*   Configuration file support (\`.contextvibes.yaml\` in project root) for:
    *   Default Git remote name (\`git.defaultRemote\`).
    *   Default Git main branch name (\`git.defaultMainBranch\`).
    *   Enabling/disabling and customizing regex patterns for branch name validation (\`validation.branchName\`).
    *   Enabling/disabling and customizing regex patterns for commit message validation (\`validation.commitMessage\`).
    *   Default AI log file name (\`logging.defaultAILogFile\`).

### Changed

*   **Architectural Refactor**: Centralized all external command execution (Git and other tools) through a new \`internal/exec.ExecutorClient\`.
    *   All \`cmd/*.go\` files (\`plan\`, \`deploy\`, \`format\`, \`test\`, \`quality\`, \`describe\`, \`kickoff\`, \`commit\`) now use this \`ExecClient\` for running external processes, replacing direct calls to \`os/exec\` or old \`internal/tools\` helpers.
    *   \`internal/git.Client\` now uses the \`exec.CommandExecutor\` interface from the \`internal/exec\` package for its underlying Git command operations.
    *   \`internal/config.FindRepoRootConfigPath\` now uses an \`ExecClient\` for \`git rev-parse\`.
*   Default AI log file name is now configurable via \`.contextvibes.yaml\` (config key: \`logging.defaultAILogFile\`, ultimate fallback: \`contextvibes_ai_trace.log\`).
*   \`cmd/kickoff.go\`: Branch naming logic updated. Now requires branches to start with \`feature/\`, \`fix/\`, \`docs/\`, or \`format/\` by default (configurable via \`.contextvibes.yaml\`). Prompts for branch name if not provided via \`--branch\` flag.
*   \`cmd/commit.go\`: Commit message validation now enforces Conventional Commits format by default (configurable via \`.contextvibes.yaml\`).

### Fixed

*   Corrected ineffective \`break\` statement in \`cmd/codemod.go\`'s \`delete_file\` operation to correctly exit the operations loop.
*   Addressed 'unused parameter: ctx' warnings in \`cmd/describe.go\` helper functions by ensuring \`ctx\` is appropriately passed to \`ExecClient\` methods or marked as used (\`_ = ctx\`) if the helper itself doesn't directly consume it.
*   Ensured all package-level variables in \`cmd/root.go\` (\`AppLogger\`, \`LoadedAppConfig\`, \`ExecClient\`, \`assumeYes\`, \`AppVersion\`, \`rootCmd\`) are correctly defined and accessible to all commands in the \`cmd\` package.
*   Updated \`.gitignore\` to explicitly ignore root-level executables and common codemod script names.

### Removed

*   \`internal/tools/exec.go\` (superseded by \`internal/exec\` package).
*   Most utility functions from \`internal/tools/git.go\` (functionality moved to \`internal/git.Client\` or uses \`os.Stat\`). \`IsGitRepo\` remains; \`CheckGitPrereqs\` is removed.
*   \`internal/git/executor.go\` (superseded by \`internal/exec.CommandExecutor\` interface).

---

## [0.0.3] - 2025-05-06

### Added

*   **\`version\` command:** Displays the current CLI version.
*   **\`test\` command:** Detects project type and runs tests.
*   Unit tests for the \`version\` command.

### Changed

*   Application version (\`AppVersion\`) set to \`0.0.3\`.
*   \`.idx/airules.md\` and \`README.md\` updated.
*   \`go.mod\`, \`go.sum\` updated for \`stretchr/testify\`.

---

## [0.0.2] - 2025-05-06

### Added

*   New \`format\` command.
*   Go project support to \`quality\` command.

### Changed

*   \`quality\` command: Go formatting check fails if files modified.
*   \`wrapup\` command: Advises on alternative workflows.
*   Documentation and internal code refinements.

### Fixed

*   Error string formatting.
*   Removed old comments.
*   Deduplicated \`.gitignore\` entries.

---

## [0.0.1] - 2025-05-06

### Added

*   **Initial Release of Context Vibes CLI.**
*   Core Functionality: \`describe\`, \`diff\`, Git Workflows, IaC Wrappers, Basic Quality Checks.
*   Project Structure: Cobra CLI, \`internal/\` packages.
*   Initial Configuration & Logging.

---

<!--
Link Definitions
-->
[Unreleased]: https://github.com/contextvibes/cli/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/contextvibes/cli/compare/v0.4.1...v0.6.0
[0.4.1]: https://github.com/contextvibes/cli/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/contextvibes/cli/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/contextvibes/cli/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/contextvibes/cli/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/contextvibes/cli/compare/v0.0.5...v0.1.1
[0.0.5]: https://github.com/contextvibes/cli/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/contextvibes/cli/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/contextvibes/cli/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/contextvibes/cli/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/contextvibes/cli/releases/tag/v0.0.1
