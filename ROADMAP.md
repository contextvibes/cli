# Context Vibes CLI - Roadmap

This document outlines the planned features and future direction for the Context Vibes CLI, considering features from its previous iteration (`context-pilot`). Priorities and features may change based on feedback and needs.

## Near Term (v0.2.x)

*   [ ] **Unit Testing:** Add comprehensive unit tests, starting with the `internal/tools` and `internal/project` packages to ensure robustness and facilitate refactoring. *(High Priority)*
*   [ ] **Configuration:** Implement support for a configuration file (e.g., `.contextvibes.yaml` or similar) to customize:
    *   Default Git remote name (`origin`).
    *   Default Git main branch name (`main`/`master`).
    *   Potentially customize `describe` include/exclude patterns or critical files.
*   [ ] **Refactor `describe`:** Break down the large `RunE` function in `cmd/describe.go` into smaller, testable helper functions within the `cmd` package for better readability and maintainability.
*   [ ] **`quality` Enhancements:** Add a check for `go mod tidy` status in Go projects as part of the `quality` command.
*   [ ] **`init` Command Flags:** Add flags to the `init` command to pass common options to underlying tools (e.g., `terraform init -upgrade`).

## Medium Term (v0.3.x / v0.4.x)

*   [ ] **`.idx/airules.md` Management (Re-evaluate Scope):**
    *   Consider re-implementing commands specifically for managing the `.idx/airules.md` file format as described in Firebase Studio documentation, if this remains a desired feature.
    *   This would likely involve:
        *   Adding back an interactive `create` command.
        *   Adding back a `maintain` command (from original `cmd/rules.go`) to validate the specific Persona/Guidelines/Context structure.
        *   Re-introducing an `internal/document` package with Markdown parsing (`goldmark`) and validation logic.
        *   Potentially adding back `ai-info` command, updated for `contextvibes`.
*   [ ] **Context File Management ("Memory"):** Explore strategies for managing the single `contextvibes.md` file, especially given `diff` overwrites it. Options include:
    *   Optional backup flag for `diff`.
    *   Timestamped output files.
    *   Alternative approaches to providing diff context alongside full context.
*   [ ] **Verbose Flag:** Introduce a global `--verbose` or `-v` flag for more detailed logging during command execution, showing underlying command output or internal steps.
*   [ ] **More Quality Tools:** Investigate and potentially add support for more linters/checkers (e.g., `golangci-lint` for Go, `mypy` for Python type checking).
*   [ ] **Richer `describe` Structure:** Explore options for displaying project structure more intelligently (e.g., using `go list` for Go projects, analyzing `pom.xml`/`package.json`, etc.).
*   [ ] **Standalone `update` Command:** Consider adding back the simpler `update` command (`git pull --rebase` with confirmation) for users who explicitly don't want the push check included in `sync`.
*   [ ] **Git Stash Integration:** Consider adding a `stash` command or integrating optional stashing into `kickoff`/`sync` workflows to handle non-clean working directories more gracefully.

## Long Term / Vision

*   [ ] **More Project Types:** Expand project detection and relevant commands (`quality`, `init`, `plan`, `deploy` steps) for other common ecosystems like Node.js (npm/yarn checks, linting), Docker (linting Dockerfiles), Rust, Java, etc.
*   [ ] **Plugin System:** Explore the possibility of a plugin architecture to allow users or teams to easily extend `contextvibes` with custom commands or project-type support without modifying the core.
*   [ ] **CI/CD Integration:** Provide clearer examples or specific flags/commands (e.g., non-interactive modes, exit codes) to facilitate the use of `contextvibes` commands within automated CI/CD pipelines.
*   [ ] **Configuration Validation:** Add a command to validate the syntax and potentially the semantics of the configuration file (once implemented).
*   [ ] **Improved AI Prompting:** Refine the structure and content of `contextvibes.md` based on feedback and best practices for different AI models.

Feedback and contributions are welcome! Please open an issue on the repository to discuss roadmap items or suggest new features.