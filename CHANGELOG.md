# Changelog

All notable changes to the **Context Vibes CLI** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

### Changed

### Removed

---
## [0.2.0] - 2025-06-08 

### Added
- **`contextvibes index` Command Enhancement:**
    - Now generates a comprehensive JSON manifest (`project_manifest.json` by default) from local Markdown files.
    - Parses rich YAML front matter including `title`, `artifactVersion`, `summary`, `usageGuidance`, `owner`, `createdDate`, `lastModifiedDate`, `defaultTargetPath`, and `tags`.
    - Derives `id` (from relative path) and `fileExtension`.
    - `lastModifiedDate` falls back to file system modification time if not in front matter.
    - `defaultTargetPath` defaults to `id.fileExtension` if not in front matter.
- **New Command: `contextvibes thea get-artifact <artifact-id>`:**
    - Fetches a specific artifact document (e.g., playbook, template) from the central THEA framework repository.
    - Uses a hardcoded default URL for the THEA manifest (`https://raw.githubusercontent.com/contextvibes/THEA/main/thea-manifest.json`) and artifacts.
    - Supports `--version` flag for version hints (maps to Git ref).
    - Supports `--output` or `-o` to specify the local save path. Defaults to a name derived from artifact metadata if `-o` is omitted.
    - Supports `--force` or `-f` to overwrite existing output files.
- **Internal `thea.Client`:**
    - Added a new internal client (`internal/thea/client.go`) responsible for interacting with the THEA framework repository.
    - Capable of fetching and parsing the remote `thea-manifest.json`.
    - Capable of fetching specific artifact content based on manifest details.
    - Includes integration tests against the live THEA repository.

### Changed
- (List any significant changes to existing functionality if applicable)
- Example: If `kickoff` was refactored to use the `thea.Client` in this release, mention it here. (Sounds like we deferred this part for now).

### Fixed
- (List any bug fixes, e.g., "Resolved `EOF` errors in `kickoff` command unit tests by mocking Stdin for collaboration preference prompts." - *IF you fixed these*)

### Removed
- (List any features removed)

## [0.1.1] - 2025-05-10

### Added
- **Enhanced `kickoff` Command (Dual Mode Functionality):**
    - **Strategic Kickoff Prompt Generation Mode** (via `contextvibes kickoff --strategic` or on first run in a new project):
        - Conducts a brief interactive session to gather initial project details (name, type, stage) and user preferences for CLI interaction styles (`ai.collaborationPreferences`).
        - Generates a comprehensive master prompt file (`STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md`) by embedding and parameterizing a detailed protocol template (from `internal/kickoff/assets/strategic_kickoff_protocol_template.md`).
        - The generated master prompt is designed for the user to take to an external AI assistant (e.g., Gemini, Claude). It instructs the AI to guide the user through a strategic kickoff checklist, request further context via `contextvibes` commands, and generate structured YAML for specified configurations.
        - Saves gathered `ai.collaborationPreferences` to `.contextvibes.yaml` after the initial setup questions (before master prompt generation and also when marking complete).
    - **Mechanism to Mark Strategic Kickoff as Complete:**
        - New `contextvibes kickoff --mark-strategic-complete` flag.
        - This command updates `.contextvibes.yaml` by setting `projectState.strategicKickoffCompleted: true` and `projectState.lastStrategicKickoffDate`. It also ensures any `ai.collaborationPreferences` gathered during a preceding `--strategic` run's setup phase are persisted.
    - **Daily Git Workflow Mode:**
        - Becomes the default for `contextvibes kickoff` (without `--strategic`) once a strategic kickoff has been marked complete for the project.
        - Includes logic for prerequisite checks (clean working directory, on main branch), prompting for/validating branch name (respecting `.contextvibes.yaml`), and executing Git operations (pull rebase, new branch, push upstream).
        - Respects the global `--yes` flag for non-interactive operation.
- **New Configuration Options in `.contextvibes.yaml`:**
    - `projectState` section:
        - `strategicKickoffCompleted` (boolean, default: `false`)
        - `lastStrategicKickoffDate` (string, RFC3339 timestamp)
    - `ai.collaborationPreferences` section:
        - `codeProvisioningStyle` (string, default: "bash_cat_eof")
        - `markdownDocsStyle` (string, default: "raw_markdown")
        - `detailedTaskMode` (string, default: "mode_b")
        - `proactiveDetailLevel` (string, default: "detailed_explanations" or "concise_unless_asked" based on taskMode)
        - `aiProactivity` (string, default: "proactive_suggestions")
- **New Internal Package `internal/kickoff`:**
    - Contains `Orchestrator` to manage all `kickoff` command logic (strategic and daily).
    - Includes embedded master kickoff protocol template (`internal/kickoff/assets/strategic_kickoff_protocol_template.md`).
- **New Documentation:**
    - `docs/PROJECT_KICKOFF_GUIDE.md` explaining the new strategic kickoff workflow.
    - Initial `DEVELOPMENT.md` and `CONTRIBUTING.md`.
- **Sample `.contextvibes.yaml`:** Added to the project root as an example.

### Changed
- **`cmd/kickoff.go`:** Major refactor to use the new `internal/kickoff.Orchestrator` and handle new flags (`--strategic`, `--mark-strategic-complete`).
- **`internal/config/config.go`:**
    - `Config` struct extended with `ProjectState` and `AISettings`.
    - `GetDefaultConfig()` updated with defaults for new AI collaboration preferences and project state.
    - `MergeWithDefaults()` updated to correctly handle merging of new nested structs (field-by-field for `AICollaborationPreferences`).
    - `UpdateAndSaveConfig()` now uses an atomic write pattern (write to temp file then rename) for improved robustness.
- **`internal/kickoff/orchestrator.go`:** `assumeYes` flag handling refined to be a field within the `Orchestrator` struct, initialized from the global flag.

### Fixed
- Various minor compilation and linter issues encountered during the development of the enhanced kickoff feature.
- Ensured `assumeYes` is correctly plumbed from `cmd/root.go` to `internal/kickoff.Orchestrator` and respected in daily Git workflows.

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
*   Default AI log file name is now configurable via `.contextvibes.yaml` (config key: `logging.defaultAILogFile`, ultimate fallback: `contextvibes_ai_trace.log`).
*   `cmd/kickoff.go`: Branch naming logic updated. Now requires branches to start with `feature/`, `fix/`, `docs/`, or `format/` by default (configurable via `.contextvibes.yaml`). Prompts for branch name if not provided via `--branch` flag.
*   `cmd/commit.go`: Commit message validation now enforces Conventional Commits format by default (configurable via `.contextvibes.yaml`).

### Fixed

*   Corrected ineffective `break` statement in `cmd/codemod.go`'s `delete_file` operation to correctly exit the operations loop.
*   Addressed 'unused parameter: ctx' warnings in `cmd/describe.go` helper functions by ensuring `ctx` is appropriately passed to `ExecClient` methods or marked as used (`_ = ctx`) if the helper itself doesn't directly consume it.
*   Ensured all package-level variables in `cmd/root.go` (`AppLogger`, `LoadedAppConfig`, `ExecClient`, `assumeYes`, `AppVersion`, `rootCmd`) are correctly defined and accessible to all commands in the `cmd` package.
*   Updated `.gitignore` to explicitly ignore root-level executables and common codemod script names.

### Removed

*   `internal/tools/exec.go` (superseded by `internal/exec` package).
*   Most utility functions from `internal/tools/git.go` (functionality moved to `internal/git.Client` or uses `os.Stat`). `IsGitRepo` remains; `CheckGitPrereqs` is removed.
*   `internal/git/executor.go` (superseded by `internal/exec.CommandExecutor` interface).

---

## [0.0.3] - 2025-05-06

### Added

*   **`version` command:** Displays the current CLI version.
*   **`test` command:** Detects project type and runs tests.
*   Unit tests for the `version` command.

### Changed

*   Application version (`AppVersion`) set to `0.0.3`.
*   `.idx/airules.md` and `README.md` updated.
*   `go.mod`, `go.sum` updated for `stretchr/testify`.

---

## [0.0.2] - 2025-05-06

### Added

*   New `format` command.
*   Go project support to `quality` command.

### Changed

*   `quality` command: Go formatting check fails if files modified.
*   `wrapup` command: Advises on alternative workflows.
*   Documentation and internal code refinements.

### Fixed

*   Error string formatting.
*   Removed old comments.
*   Deduplicated `.gitignore` entries.

---

## [0.0.1] - 2025-05-06

### Added

*   **Initial Release of Context Vibes CLI.**
*   Core Functionality: `describe`, `diff`, Git Workflows, IaC Wrappers, Basic Quality Checks.
*   Project Structure: Cobra CLI, `internal/` packages.
*   Initial Configuration & Logging.

---

<!--
Link Definitions
-->
[Unreleased]: https://github.com/contextvibes/cli/compare/v0.1.0...HEAD
[0.1.1]: https://github.com/contextvibes/cli/compare/v0.0.5...v0.1.1
[0.0.5]: https://github.com/contextvibes/cli/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/contextvibes/cli/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/contextvibes/cli/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/contextvibes/cli/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/contextvibes/cli/releases/tag/v0.0.1
