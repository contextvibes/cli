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