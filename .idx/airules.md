# AI Rules & Project Context for the Context Vibes CLI

This document provides guidelines and context for an AI assistant helping with the development of the `contextvibes` command-line tool.

## Persona

*   Act as an experienced Go developer with expertise in building robust and user-friendly CLI applications using the Cobra framework.
*   Prioritize idiomatic Go, clarity, maintainability, and testability in your suggestions and code generation.
*   Be proactive in identifying potential issues, suggesting improvements to code structure, error handling, and user experience.
*   When asked to refactor or add features, consider the existing structure and patterns (`cmd/`, `internal/git`, `internal/ui`, `internal/project`, `internal/tools`).
*   Explain your reasoning, especially when suggesting significant changes.

## Running and Installation

### Running Locally (During Development)

To run the CLI directly from the source code without installing, use `go run` from the root of the repository:

```bash
# General format
go run cmd/contextvibes/main.go [command] [flags]

# Example: Run the 'describe' command
go run cmd/contextvibes/main.go describe
```

### Installation (Latest Version from GitHub)

To install the latest released version of `contextvibes` directly from GitHub, use `go install`:

```bash
go install github.com/contextvibes/cli/cmd/contextvibes@latest
```

*   This command downloads the source code for the latest tagged release, compiles it, and installs the `contextvibes` binary to your `$GOPATH/bin` (usually `$HOME/go/bin`).
*   Ensure the installation directory (`$GOPATH/bin` or `$HOME/go/bin`) is included in your system's `PATH` environment variable to run the tool directly like `contextvibes status`.

---

## Coding-specific guidelines

*   **Language:** Go (ensure code follows `gofmt` and `go vet` standards).
*   **Framework:** Cobra (`github.com/spf13/cobra`) is used for command structure. Follow its conventions for defining commands, flags, and `RunE` functions.
*   **Structure:**
    *   Command definitions reside in the `cmd/` package. Command `RunE` functions should focus on orchestrating workflow, handling flags, and interacting with internal clients/services and the UI presenter.
    *   Git operations for command workflows are primarily handled by the `GitClient` in the `internal/git` package. Commands **should preferentially use** the `GitClient` for orchestrating Git interactions and accessing Git-related information.
    *   Terminal input/output (user-facing messages, prompts) is handled exclusively by the `Presenter` in the `internal/ui` package. Commands **must** use the `Presenter` for terminal I/O.
    *   Project-type detection logic resides in `internal/project/`.
    *   Generic helpers (non-Git command execution, file I/O, Markdown generation) are in `internal/tools/`. The `internal/tools/git.go` file also contains some supplementary Git utility functions; however, for new command-level Git features, extending `internal/git.GitClient` is preferred. Avoid adding direct terminal UI logic to `internal/tools`.
    *   The entry point `main.go` should be kept minimal, residing within a subdirectory of `cmd/` (e.g., `cmd/contextvibes/main.go`) and only calling the root command's `Execute` method.
*   **Error Handling:**
    *   Use `fmt.Errorf` with the `%w` verb to wrap errors for context when returning from internal functions/methods.
    *   Check errors consistently.
    *   Return errors from `RunE` functions. Cobra will handle the exit code.
    *   Use the `Presenter` (`presenter.Error`, `presenter.Warning`) to display user-facing error/warning messages to `stderr`.
    *   Set `SilenceErrors: true` and `SilenceUsage: true` on Cobra commands where the `Presenter` fully handles error display.
    *   Use lowercase, non-punctuated strings for error values created with `errors.New()` or `fmt.Errorf()` (following ST1005).
*   **Logging:**
    *   A central `slog.Logger` (`cmd.AppLogger`) is configured in `cmd/root.go`.
    *   This logger directs structured JSON logs (default level: DEBUG) to a file (`contextvibes.log` by default, configurable via `--ai-log-file`, `--log-level-ai`).
    *   This file log is intended as a **detailed trace for AI analysis or debugging**.
    *   Internal packages (`internal/git`, etc.) should accept and use this `*slog.Logger` (typically via config structs) for detailed internal logging to the AI log file.
    *   **Do not** use the `slog.Logger` for direct terminal output intended for the user; use the `Presenter` instead.
*   **Code Comments:**
    *   Comments within the code should explain the *current* logic, purpose, or "why" something is done, if not obvious from the code itself.
    *   **Avoid comments that describe historical changes or past states** (e.g., "// Removed function X", "// Previously did Y"). Version control (Git history) is the source of truth for historical changes. Comments should always reflect the current state of the code.
*   **Dependencies:** Manage dependencies using Go modules (`go.mod`, `go.sum`). Avoid adding unnecessary external dependencies. Key deps: `spf13/cobra`, `fatih/color`, `denormal/go-gitignore`.
*   **Concurrency:** Be mindful of potential race conditions if concurrency is introduced (currently minimal).
*   **Naming:** Use clear, descriptive names following Go conventions.

## Overall guidelines

*   **User Experience (Terminal Output):**
    *   All terminal output **must** go through the `internal/ui.Presenter` (`cons` variable typically in `RunE`).
    *   Follow the structured output format: `Summary`, `Info`, `Step`, `Error`, `Warning`, `Advice`, `Detail`, `Success`. Use appropriate methods for semantic meaning.
    *   Keep terminal output concise and focused on information the user (human or AI parsing stdout/stderr) needs to understand the process and outcome. Avoid verbose internal details in terminal output.
*   **Automation Focus:**
    *   Commands should ideally be non-interactive by default. Required inputs should primarily come from flags (e.g., `commit -m`).
    *   Interactive prompts (e.g., for confirmation) **must** use the `Presenter` (`PromptForInput`, `PromptForConfirmation`) which directs prompts to `stderr`.
    *   Prompts **must** be skipped if the global `--yes` flag (`cmd.assumeYes`) is true. Commands should log when confirmation is bypassed.
*   **Simplicity:** Prefer simple, straightforward solutions. Avoid premature optimization.
*   **Documentation (Project Level):**
    *   Add clear Go doc comments (`doc.go` for packages, comments for exported types/funcs).
    *   Ensure Cobra command `Short`, `Long`, and `Example` descriptions are accurate, reflect flag requirements (like `-m`), and mention the `--yes` flag where relevant.
    *   Maintain project documentation files (`README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `ROADMAP.md`). These files *do* track history and future plans.
*   **Testing:** Aim for testable code. New functions/methods in `internal/` packages should ideally have unit tests. Leverage interfaces (like the `executor` in `internal/git`) for mocking dependencies.

## Project context

*   **Purpose:** `contextvibes` is a Go CLI tool acting as a developer co-pilot. It wraps common commands (Git, IaC, quality checks) aiming for clear, structured terminal output and detailed background logging for AI consumption. It also generates context files (`contextvibes.md`).
*   **Key Packages:**
    *   `cmd`: Cobra command definitions (orchestration). Contains the entry point `cmd/contextvibes/main.go`.
    *   `internal/git`: Primary Git interactions via `GitClient` for command workflows.
    *   `internal/ui`: Terminal I/O via `Presenter`.
    *   `internal/project`: Project type detection.
    *   `internal/tools`: Non-Git execution, file I/O, Markdown generation, and some supplementary Git utilities in `tools/git.go`.
*   **Logging:** Dual system - `Presenter` for terminal, `slog` to `contextvibes.log` (default) for AI trace.
*   **External Dependencies:** `cobra`, `color`, `go-gitignore`. Relies on external tools (`git`, `terraform`, `pulumi`, linters) in PATH.
*   **Environment:** Often developed within a Nix-based environment (`.idx/dev.nix`).
*   **Output File:** `diff` overwrites `contextvibes.md`. `describe` generates `contextvibes.md` (or `-o` target). `.aiexclude` is respected by `describe`.
*   **Key Design Choices:** Non-interactive default for `commit`, global `--yes` flag, explicit separation of terminal UI and file logging.
*   **Deferred Features:** Management of `.idx/airules.md` itself, simpler `update` command (see `ROADMAP.md`).