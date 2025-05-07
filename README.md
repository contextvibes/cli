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
    *   Produces structured terminal output (using `SUMMARY:`, `INFO:`, `ERROR:`, `ADVICE:`, and status prefixes like `+`, `-`, `~`, `!`) suitable for human review or direct AI parsing.
    *   Generates a detailed JSON trace log (default: `contextvibes_ai_trace.log`, configurable) for deeper AI analysis or debugging, separate from terminal output.
*   **Clarity & Safety:** Uses distinct output formats for different information types and requires confirmation for state-changing operations (unless `--yes` is specified).
*   **Configurability:** Supports a `.contextvibes.yaml` file in the project root for customizing default behaviors like Git branch/remote names, validation patterns, and default log file names.

## Key Features

*   **AI Context Generation:**
    *   `describe`: Gathers project details (env, git status, files, user prompt) into `contextvibes.md`. Supports `-o` for output file customization. Respects `.gitignore` and `.aiexclude`.
    *   `diff`: Summarizes pending Git changes (staged, unstaged, untracked), **overwriting** `contextvibes.md`.
*   **Git Workflow Automation (Configurable branch/commit rules):**
    *   `kickoff`: Start-of-work routine (update main, create/switch new work branch). Requires confirmation. Branch name validation configurable.
    *   `commit`: Stages all changes and commits locally. **Requires `-m <message>` flag.** Requires confirmation. Commit message format validation configurable.
    *   `sync`: Updates local branch from remote (`pull --rebase`) and pushes if ahead. Requires clean state and confirmation.
    *   `wrapup`: End-of-day routine (stages all, commits with default message if needed, pushes). Requires confirmation.
    *   `status`: Shows concise `git status --short` output.
*   **Infrastructure as Code (IaC) Wrappers:**
    *   `plan`: Run `terraform plan -out=tfplan.out` or `pulumi preview`.
    *   `deploy`: Run `terraform apply tfplan.out` or `pulumi up`. Requires confirmation.
    *   `init`: Run `terraform init` (primarily for Terraform projects).
*   **Code Quality & Formatting:**
    *   `quality`: Run formatters (in check mode), validators, and linters (Terraform, Python, Go support included).
    *   `format`: Apply code formatting (`go fmt`, `terraform fmt`, `isort`, `black`) modifying files in place.
*   **Project Testing & Versioning:**
    *   `test`: Runs project-specific tests (e.g., `go test ./...`, `pytest`). Arguments are passed to the underlying test runner.
    *   `version`: Displays the CLI version.
*   **Code Modification:**
    *   `codemod`: Applies programmatic code modifications or deletions from a JSON script (e.g., `codemod.json`).

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

**Dependencies:** Relies on external tools being in your `PATH`: `git`, and depending on project type and commands used: `terraform`, `pulumi`, `tflint`, `isort`, `black`, `flake8`, `go` (for Go project commands), `python` (for Python project commands).

## Usage

```bash
contextvibes [command] --help  # General help or help for a specific command
contextvibes [command] [flags] # Run a command
```

**Common Flags:**

*   `-y`, `--yes`: Assume 'yes' to all confirmation prompts. Useful for scripts and automation.
*   `--ai-log-file <path>`: Specify a file path for the detailed AI JSON log (overrides config default).
*   `--log-level-ai <level>`: Set the minimum level for the AI log file (debug, info, warn, error; default: `debug`).

**Examples:**

```bash
# Start a new feature branch (prompts for name if not provided via --branch)
contextvibes kickoff --branch feature/add-user-auth

# Describe the project for an AI (prompts for task description)
contextvibes describe -o my_context.md

# Show pending Git changes (overwrites contextvibes.md)
contextvibes diff

# Apply code formatting
contextvibes format

# Check code quality
contextvibes quality

# Run project tests (e.g., for a Go project, passing -v flag)
contextvibes test -v

# Plan infrastructure changes
contextvibes plan

# Commit work (message required, interactive confirmation)
contextvibes commit -m "feat(auth): Implement OTP login"

# Commit work (non-interactive)
contextvibes commit -m "fix(api): Correct typo in user model" -y

# Sync with remote (requires confirmation)
contextvibes sync

# Sync non-interactively
contextvibes sync -y

# End your day (requires confirmation)
contextvibes wrapup

# End your day non-interactively
contextvibes wrapup -y

# Display CLI version
contextvibes version

# View command options
contextvibes commit --help

# Apply programmatic changes from a script (default: codemod.json)
contextvibes codemod
contextvibes codemod --script ./changes.json
```

## Terminal Output vs. AI Log File

Context Vibes uses two distinct output mechanisms:

1.  **Terminal Output (stdout/stderr):**
    *   Designed for **human readability** and **potential AI parsing** of the command's outcome and context.
    *   Uses structured blocks (`SUMMARY:`, `INFO:`, `ERROR:`, `ADVICE:`) and status prefixes (`+`, `-`, `~`, `!`).
    *   `stdout` generally contains status information, results, and advice.
    *   `stderr` contains operational errors and interactive prompts (like confirmation).
2.  **AI Log File (JSON):**
    *   Written to `contextvibes_ai_trace.log` by default (configurable via `.contextvibes.yaml` or `--ai-log-file`).
    *   Contains a **detailed, structured trace** of the command's internal execution steps, parameters, and outcomes using `slog`.
    *   Intended for deeper analysis by AI agents or for debugging by humans.
    *   Log level controlled by `--log-level-ai` (defaults to `debug`).

## Configuration (`.contextvibes.yaml`)

`contextvibes` can be configured using a `.contextvibes.yaml` file placed in the root of your project. This allows for project-specific settings that can be version-controlled. If the file is not present, the CLI uses sensible built-in defaults.

**Example `.contextvibes.yaml`:**

```yaml
# Git settings
git:
  defaultRemote: origin
  defaultMainBranch: main # or master, develop, etc.

# Logging settings
logging:
  defaultAILogFile: "logs/contextvibes_ai.log" # Path relative to project root

# Validation rules
validation:
  branchName:
    enable: true  # Default is true. Set to false to disable validation.
    # pattern: "^(feature|fix|chore|task)/[A-Z]+-[0-9]+-.*$" # Example custom pattern
  commitMessage:
    enable: true # Default is true. Set to false to disable validation.
    # pattern: "^(feat|fix|chore|docs)!?(\\(.+\\))?: .+" # Example custom pattern
```

*   If a `pattern` is not specified for an enabled validation rule, the CLI's built-in default pattern will be used.
*   The `--ai-log-file` CLI flag will always override the `logging.defaultAILogFile` setting.

## Note on AI Context Files (`contextvibes.md`, `.idx/airules.md`, `.aiexclude`)

*   **`contextvibes.md` (Generated by `describe`/`diff`):** Dynamic snapshot of project state/diff + user prompt. Use this as the **input prompt** for external AI models.
*   **`.idx/airules.md` (IDE Config):** *Static configuration* for IDE-native AI features (e.g., Project IDX). `contextvibes` currently reads but does not generate this (see Roadmap).
*   **`.aiexclude` (Exclusion Rules):** User-defined rules (like `.gitignore`) to exclude files/dirs from `describe`.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Key areas include adding unit tests, refactoring remaining non-Git commands, and updating documentation.

## Code of Conduct

Act professionally and respectfully. Be kind, considerate, and welcoming. Harassment or exclusionary behavior will not be tolerated.

## Related Files

*   `CHANGELOG.md`: Tracks notable changes.
*   `CONTRIBUTING.md`: Contribution guidelines, known issues/TODOs.
*   `ROADMAP.md`: Future plans.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.