# Changelog

All notable changes to the **Context Vibes CLI** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.0.1] - YYYY-MM-DD <!-- Set the actual release date -->

### Added

*   **Initial Release of Context Vibes CLI.**
*   **Core Functionality:**
    *   **AI Context Generation:**
        *   `describe`: Generates `contextvibes.md` with project context (environment, git status, files, user prompt). Supports output file customization (`-o`), respects `.gitignore` and `.aiexclude`.
        *   `diff`: Generates a diff summary (staged, unstaged, untracked) **overwriting** `contextvibes.md`.
    *   **Git Workflow Automation:**
        *   `kickoff`: Prepares a daily `dev-YYYY-MM-DD` branch (requires clean main, updates main, creates/switches branch, requires confirmation).
        *   `commit`: Stages all changes and commits locally (**requires message via `-m` flag**, requires confirmation).
        *   `sync`: Safely synchronizes the current branch with its remote (checks clean, pulls rebase, pushes if ahead, requires confirmation).
        *   `wrapup`: Performs end-of-day routine: stages all changes, commits with a default message (`chore: Automated wrapup commit`) if needed, pushes. Requires confirmation. Includes advice about using alternative commands (`quality`, `commit -m`, `sync`) for more control.
        *   `status`: Displays concise `git status --short` output.
        *   `format`: Applies standard code formatting (`go fmt`, `terraform fmt`, `isort`, `black`) modifying files in place.
    *   **Infrastructure as Code (IaC) Wrappers:**
        *   `plan`: Detects project type (Terraform/Pulumi) and runs the corresponding plan/preview command (`terraform plan -out=tfplan.out` / `pulumi preview`).
        *   `deploy`: Detects project type (Terraform/Pulumi) and runs the corresponding deploy command (`terraform apply tfplan.out` / `pulumi up`, requires confirmation).
        *   `init`: Primarily for Terraform; runs `terraform init`.
    *   **Code Quality:**
        *   `quality`: Detects project type (Terraform/Python/Go) and runs relevant formatters (in check mode), validators, and linters.
            *   Terraform: `terraform fmt -check`, `terraform validate`, `tflint`.
            *   Python: `isort --check`, `black --check`, `flake8`.
            *   Go: `go fmt` (note: runs format, fails if files *were* modified), `go vet`, `go mod tidy`.
            *   Reports critical failures (format checks, validators, `go vet`, `go mod tidy`) as errors causing command failure.
            *   Reports linter issues (`tflint`, `flake8`) as warnings.
*   **Project Structure:**
    *   Uses Cobra CLI framework (`cmd/`).
    *   Separation of concerns into internal packages:
        *   `internal/git`: Primary Git interactions via `GitClient`.
        *   `internal/ui`: Terminal I/O via `Presenter`.
        *   `internal/project`: Project type detection (Terraform, Pulumi, Go, Python).
        *   `internal/tools`: Generic command execution (`exec.go`), file I/O (`io.go`), Markdown helpers (`markdown.go`), and supplementary Git utilities (`git.go`).
*   **Configuration & Rules:**
    *   Uses `.idx/airules.md` for AI assistant guidelines (read by `describe`, not managed by tool).
    *   Supports `.aiexclude` for `describe` command.
*   **Logging:** Dual system: Structured terminal output via `Presenter`; detailed JSON trace logging to file (`contextvibes.log` by default) via `slog`.
*   **Core Dependencies:** Go, Cobra, `fatih/color`, `denormal/go-gitignore`. Relies on external tools (`git`, `terraform`, `pulumi`, `tflint`, etc.) being in PATH.

---

<!--
Link Definitions - Add the new one when tagging
-->
[0.0.1]: https://github.com/contextvibes/cli/.../tag/v0.0.1 <!-- Adjust URL and tag path -->