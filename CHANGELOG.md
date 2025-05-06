# Changelog

All notable changes to the **Context Vibes CLI** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
[0.0.3]: https://github.com/contextvibes/cli/.../compare/v0.0.2...v0.0.3 <!-- Adjust URL and tags -->
[0.0.2]: https://github.com/contextvibes/cli/.../compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/contextvibes/cli/.../tag/v0.0.1