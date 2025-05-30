# Context Vibes CLI - Roadmap

This document outlines the planned features and future direction for the Context Vibes CLI. Priorities and features may change based on feedback and community contributions. This roadmap aims to guide development efforts and inform potential contributors about areas where help is welcome.

## Near Term (Next 1-2 Minor Releases, including post-v0.1.0)

*   [ ] **Unit Tests for Enhanced Kickoff (`internal/kickoff`):** *(High Priority post v0.1.0)*
    *   Implement comprehensive unit tests for `internal/kickoff.Orchestrator` methods.
    *   This includes mocking `ui.Presenter`, `git.ClientInterface` (may require refactoring Orchestrator to use a Git client interface), and `config.Config` interactions.
    *   Focus on testing:
        *   Mode selection logic in `ExecuteKickoff` (strategic vs. daily).
        *   `runCollaborationSetup`, `runInitialInfoGathering`, `runTechnicalReadinessInquiry` with mocked presenter inputs and config state verification.
        *   `generateMasterKickoffPromptText` for correct template execution and parameter substitution.
        *   `generateCollaborationPrefsYAML` output.
        *   `MarkStrategicKickoffComplete` for correct config updates and save calls.
        *   Detailed scenarios for `executeDailyKickoff` (mocking Git operations for success/failure, branch validation, `assumeYes` behavior).
*   [ ] **Comprehensive Unit Testing (General):** *(High Priority, ongoing)*
    *   Expand unit test coverage significantly. Priority areas:
        *   `internal/git.GitClient` methods (ideally refactoring to use an interface for easier mocking).
        *   `internal/exec` package components.
        *   Further tests for `internal/config` package logic (edge cases for `LoadConfig`, `UpdateAndSaveConfig` failures, `MergeWithDefaults` with complex overrides).
        *   `internal/tools` utilities.
        *   `internal/project` package detection logic.
    *   Goal: Ensure robustness, facilitate safe refactoring, and improve maintainability.
*   [ ] **Refactor Orchestrator Dependencies to Interfaces:** *(Medium Priority - consider for v0.2.0 or if kickoff unit testing becomes too complex)*
    *   Modify `internal/kickoff.Orchestrator` and its constructor `NewOrchestrator` to accept interfaces for `ui.Presenter` and `git.GitClient` instead of concrete types. This will significantly simplify unit testing, especially for `executeDailyKickoff`.
*   [ ] **Implement Configurable Commit Message Validation in `cmd/commit.go` (Finalize):**
    *   Ensure the `commit` command fully utilizes the `validation.commitMessage.enable` and `validation.commitMessage.pattern` settings from `.contextvibes.yaml` as per v0.0.5 changelog (verify this is complete and robust).
*   [ ] **Refactor `describe` Command:**
    *   Break down the large `RunE` function in `cmd/describe.go` into smaller, more manageable, and testable helper functions within the `cmd` package or a new `internal/describe_logic` package. This will improve readability and maintainability.
*   [ ] **Enhance `quality` for `go mod tidy`:**
    *   Modify the `go mod tidy` check in the `quality` command to detect if `go.mod` or `go.sum` files *were modified* by `go mod tidy`. If modifications occur, the check should fail (similar to how `go fmt` non-compliance is handled), prompting the user to commit the changes.
*   [ ] **`init` Command Flag Enhancements:**
    *   Add flags to the `init` command (especially for Terraform) to pass common options to the underlying initialization tools (e.g., `terraform init -upgrade`, `terraform init -reconfigure`).
*   [ ] **Generate Kickoff Summary Document:**
    *   Implement the logic in `internal/kickoff.Orchestrator` (part of `executeStrategicKickoffGeneration`) to generate a structured Markdown summary document based on the (simulated for now) Q&A from the strategic kickoff. This was a TODO in the `v0.1.0` plan.

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
        *   Go: `golangci-lint` (ensure robust integration if not already fully covered by `quality` command).
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
*   [ ] **Direct AI Integration (Optional Feature):**
    *   Investigate safely and effectively integrating direct calls to LLM APIs for commands like `ask` or `plan` (similar to `vibe-tools` model), making ContextVibes a more active AI agent.
    *   This would require robust API key management, user consent for data transfer, and clear indication of when external AI is being called.
*   [ ] **Plugin System or Extensibility:**
    *   Explore architectural changes to allow users or teams to more easily extend `contextvibes` with custom commands, project-type specific logic, or quality tool integrations without modifying the core CLI codebase.
*   [ ] **CI/CD Integration Enhancements:**
    *   Provide clearer examples, documentation, or specific flags/output modes (e.g., machine-readable output for some commands) to facilitate the robust use of `contextvibes` commands within automated CI/CD pipelines.
*   [ ] **Configuration File Validation:**
    *   Add a dedicated command (e.g., `contextvibes config validate`) to check the syntax and potentially the semantic correctness of the `.contextvibes.yaml` file.
*   [ ] **Improved AI Prompting & Context (`contextvibes.md` and Master Kickoff Prompt):**
    *   Continuously refine the structure, content, and verbosity of the `contextvibes.md` file and the generated `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md` based on evolving best practices for prompting various AI models and user feedback.

---

This roadmap is a living document. Feedback, suggestions, and contributions are highly welcome! Please open an issue on the GitHub repository to discuss roadmap items or propose new features.