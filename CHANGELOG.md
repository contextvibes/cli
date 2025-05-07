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