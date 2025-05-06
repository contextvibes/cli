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

Context Vibes is a command-line tool designed to streamline common development tasks and generate context for AI assistants. It provides consistent wrappers for Git workflows, Infrastructure as Code (IaC) operations, and code quality checks, focusing on clear, structured terminal output and detailed background logging.

## Why Context Vibes?

*   **Consistency:** Provides a unified interface and terminal output style for frequent actions (`commit`, `sync`, `deploy`, etc.).
*   **Automation:** Simplifies multi-step processes and provides non-interactive options via the global `--yes` flag. Designed for use in scripts or by AI agents.
*   **AI Integration:**
    *   Generates a `contextvibes.md` context file (`describe`, `diff`) suitable for AI prompts.
    *   Produces structured terminal output (using `SUMMARY:`, `INFO:`, `ERROR:`, `ADVICE:`, and status prefixes like `+`, `-`, `~`, `!`) suitable for human review or direct AI parsing.
    *   Generates a detailed JSON trace log (`contextvibes.log` by default) for deeper AI analysis or debugging, separate from terminal output.
*   **Clarity & Safety:** Uses distinct output formats for different information types and requires confirmation for state-changing operations (unless `--yes` is specified).

## Key Features

*   **AI Context Generation:**
    *   `describe`: Gathers project details (env, git status, files, user prompt) into `contextvibes.md`. Supports `-o` for output file customization. Respects `.gitignore` and `.aiexclude`.
    *   `diff`: Summarizes pending Git changes (staged, unstaged, untracked), **overwriting** `contextvibes.md`.
*   **Git Workflow Automation:**
    *   `kickoff`: Start-of-day routine (update main, create/switch daily dev branch). Requires confirmation.
    *   `commit`: Stages all changes and commits locally. **Requires `-m <message>` flag.** Requires confirmation.
    *   `sync`: Updates local branch from remote (`pull --rebase`) and pushes if ahead. Requires clean state and confirmation.
    *   `wrapup`: End-of-day routine (stages all, commits with default message if needed, pushes). Requires confirmation.
    *   `status`: Shows concise `git status --short` output.
*   **Infrastructure as Code (IaC) Wrappers:**
    *   `plan`: Run `terraform plan -out=tfplan.out` or `pulumi preview`.
    *   `deploy`: Run `terraform apply tfplan.out` or `pulumi up`. Requires confirmation.
    *   `init`: Run `terraform init` (primarily for Terraform projects).
*   **Code Quality:**
    *   `quality`: Run formatters and linters (Terraform, Python support included).

## Installation

Ensure you have Go (`1.24` or later recommended) and Git installed.

1.  **Install using `go install`:**
    ```bash
    go install github.com/contextvibes/cli/cmd/contextvibes@latest
    ```
    Installs to `$GOPATH/bin` or `$HOME/go/bin`.

2.  **Ensure Installation Directory is in your `PATH`:**
    ```bash
    # Add one of these to your shell profile (.bashrc, .zshrc, etc.)
    export PATH=$(go env GOPATH)/bin:$PATH
    # Or: export PATH=$HOME/go/bin:$PATH
    ```
    Restart your shell or source the profile (`source ~/.bashrc`).

**(Alternative) Installation via Releases:** Download from [GitHub Releases](https://github.com/contextvibes/cli/releases) (*Adjust URL*), make executable (`chmod +x`), move to a directory in your `PATH`.

**Dependencies:** Relies on external tools being in your `PATH`: `git`, `terraform`, `pulumi`, `tflint`, `isort`, `black`, `flake8` (depending on project type and commands used).

## Usage

```bash
contextvibes [command] --help  # General help or help for a specific command
contextvibes [command] [flags] # Run a command
```

**Common Flags:**

*   `-y`, `--yes`: Assume 'yes' to all confirmation prompts. Useful for scripts and automation.
*   `--ai-log-file <path>`: Specify a file path for the detailed AI JSON log (default: `contextvibes.log`).
*   `--log-level-ai <level>`: Set the minimum level for the AI log file (debug, info, warn, error; default: `debug`).

**Examples:**

```bash
# Start your day (interactive confirmation)
contextvibes kickoff

# Describe the project for an AI (prompts for task description)
contextvibes describe -o my_context.md

# Show pending Git changes (overwrites contextvibes.md)
contextvibes diff

# Check code quality
contextvibes quality

# Plan infrastructure changes
contextvibes plan

# Commit work (message required, interactive confirmation)
contextvibes commit -m "feat: Implement user login"

# Commit work (non-interactive)
contextvibes commit -m "fix: Correct typo in docs" -y

# Sync with remote (requires confirmation)
contextvibes sync

# Sync non-interactively
contextvibes sync -y

# End your day (requires confirmation)
contextvibes wrapup

# End your day non-interactively
contextvibes wrapup -y

# View command options
contextvibes commit --help
```

## Terminal Output vs. AI Log File

Context Vibes uses two distinct output mechanisms:

1.  **Terminal Output (stdout/stderr):**
    *   Designed for **human readability** and **potential AI parsing** of the command's outcome and context.
    *   Uses structured blocks (`SUMMARY:`, `INFO:`, `ERROR:`, `ADVICE:`) and status prefixes (`+`, `-`, `~`, `!`).
    *   `stdout` generally contains status information, results, and advice.
    *   `stderr` contains operational errors and interactive prompts (like confirmation).
2.  **AI Log File (JSON):**
    *   Written to `contextvibes.log` by default (configurable via `--ai-log-file`).
    *   Contains a **detailed, structured trace** of the command's internal execution steps, parameters, and outcomes using `slog`.
    *   Intended for deeper analysis by AI agents or for debugging by humans.
    *   Log level controlled by `--log-level-ai` (defaults to `debug`).

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