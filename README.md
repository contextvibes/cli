# Context Vibes CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/contextvibes/cli)](https://goreportcard.com/report/github.com/contextvibes/cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<!-- Open in Firebase Studio Button -->
<a href="https://studio.firebase.google.com/import?url=https%3A%2F%2Fgithub.com%2Fcontextvibes%2Fcli">
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
    *   The `kickoff --strategic` command generates a master prompt file to guide an AI-facilitated strategic project kickoff.
    *   Produces structured terminal output suitable for human review or direct AI parsing.
    *   Generates a detailed JSON trace log (default: `contextvibes_ai_trace.log`, configurable) for deeper AI analysis or debugging.
*   **Clarity & Safety:** Uses distinct output formats and requires confirmation for state-changing operations (unless `--yes` is specified).
*   **Configurability:** Supports a `.contextvibes.yaml` file for customizing default behaviors (Git, validation rules, logging, AI interaction preferences). See the [Configuration Reference](docs/reference/configuration_reference.md) for details.

## Key Features

*   **AI Context Generation:** `describe`, `diff`.
*   **Enhanced Git Workflow Automation:**
    *   `kickoff`: Dual-mode for AI-assisted strategic project kickoff prompt generation OR daily development branch creation.
    *   `commit`, `sync`, `wrapup`, `status` (Configurable branch/commit rules).
*   **Infrastructure as Code (IaC) Wrappers:** `plan`, `deploy`, `init` (Terraform/Pulumi).
*   **Code Quality & Formatting:** `quality`, `format` (Go, Python, Terraform).
*   **Project Testing & Versioning:** `test`, `version`.
*   **Code Modification:** `codemod` (from JSON scripts).

*(For detailed information on each command, see the [Command Reference](docs/reference/command_reference.md).)*

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

**Dependencies:** Relies on external tools being in your `PATH`: `git`, and potentially `terraform`, `pulumi`, `tflint`, `isort`, `black`, `flake8`, `python`.

## Usage

```bash
contextvibes [command] --help  # General help or help for a specific command
contextvibes [command] [flags] # Run a command
```

**Common Flags:**

*   `-y`, `--yes`: Assume 'yes' to all confirmation prompts.
*   `--ai-log-file <path>`: Specify a path for the detailed AI JSON log.
*   `--log-level-ai <level>`: Set the minimum level for the AI log file (debug, info, warn, error).

*(See the [Command Reference](docs/reference/command_reference.md) for all commands and flags.)*

**Examples:**

```bash
# Initiate a strategic project kickoff (generates a master prompt for your AI)
contextvibes kickoff --strategic

# (After AI session and creating summary) Mark strategic kickoff as complete
contextvibes kickoff --mark-strategic-complete

# Describe the project for an AI (prompts for task description)
contextvibes describe -o my_context.md

# Run project tests (e.g., for a Go project, passing -v flag)
contextvibes test -v
```

## Documentation

Our documentation is organized to help you find information quickly.

*   **Project Overview:** [README.md](README.md) (this file)
*   **Changelog:** [CHANGELOG.md](CHANGELOG.md)

### Guides
*   **User Manual:** [docs/guides/user_manual.md](docs/guides/user_manual.md)
*   **Project Kickoff Guide:** [docs/guides/project_kickoff_guide.md](docs/guides/project_kickoff_guide.md)

### Reference
*   **Command Reference:** [docs/reference/command_reference.md](docs/reference/command_reference.md) (Stub)
*   **Configuration Reference:** [docs/reference/configuration_reference.md](docs/reference/configuration_reference.md)

### Development & Contributing
*   **Local Development Guide:** [docs/development/development_guide.md](docs/development/development_guide.md)
*   **Contributing Guide:** [docs/development/CONTRIBUTING.md](docs/development/CONTRIBUTING.md)
*   **Project Roadmap:** [docs/development/roadmap.md](docs/development/roadmap.md)

## Important: Ignoring Generated Files

It is strongly recommended to add generated files like `contextvibes.md`, `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md`, `contextvibes_ai_trace.log`, `*.log`, and `tfplan.out` to your project's `.gitignore` file.

## Terminal Output vs. AI Log File

Context Vibes uses two distinct output mechanisms:

1.  **Terminal Output (stdout/stderr):** For human readability and high-level status/errors.
2.  **AI Log File (JSON):** Written to `contextvibes_ai_trace.log` by default (configurable).

## Code of Conduct

Act professionally and respectfully. Be kind, considerate, and welcoming. Harassment or exclusionary behavior will not be tolerated.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.