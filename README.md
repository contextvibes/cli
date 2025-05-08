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
    *   Produces structured terminal output suitable for human review or direct AI parsing.
    *   Generates a detailed JSON trace log (default: `contextvibes_ai_trace.log`, configurable) for deeper AI analysis or debugging.
*   **Clarity & Safety:** Uses distinct output formats and requires confirmation for state-changing operations (unless `--yes` is specified).
*   **Configurability:** Supports a `.contextvibes.yaml` file for customizing default behaviors. See the [Configuration Reference](docs/CONFIGURATION_REFERENCE.md) for details.

## Key Features

*   **AI Context Generation:** `describe`, `diff`
*   **Git Workflow Automation:** `kickoff`, `commit`, `sync`, `wrapup`, `status` (Configurable branch/commit rules)
*   **Infrastructure as Code (IaC) Wrappers:** `plan`, `deploy`, `init`
*   **Code Quality & Formatting:** `quality`, `format`
*   **Project Testing & Versioning:** `test`, `version`
*   **Code Modification:** `codemod`

*(For detailed information on each command, see the [Command Reference](docs/COMMAND_REFERENCE.md).)*

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

*(See the [Command Reference](docs/COMMAND_REFERENCE.md) for all commands and flags.)*

**Examples:**

```bash
# Start a new feature branch (prompts for name if needed)
contextvibes kickoff --branch feature/add-user-auth

# Describe the project for an AI (prompts for task description)
contextvibes describe -o my_context.md

# Apply code formatting
contextvibes format

# Check code quality
contextvibes quality

# Run project tests (e.g., for a Go project, passing -v flag)
contextvibes test -v

# Commit work (message required, interactive confirmation)
contextvibes commit -m "feat(auth): Implement OTP login"

# Sync non-interactively
contextvibes sync -y

# Display CLI version
contextvibes version

# Apply programmatic changes from a script
contextvibes codemod --script ./changes.json```

## Documentation

*   **[Overview & Installation](README.md):** (This file) High-level features and setup.
*   **[Command Reference](docs/COMMAND_REFERENCE.md):** Detailed syntax, flags, examples, and exit codes for every command. **Use this for specific command usage.**
*   **[Configuration Reference](docs/CONFIGURATION_REFERENCE.md):** Full details on configuring the CLI via `.contextvibes.yaml`. **Use this to customize behavior.**
*   **[Contributing Guidelines](CONTRIBUTING.md):** How to contribute code, report issues, and set up a development environment.
*   **[Changelog](CHANGELOG.md):** History of notable changes in each release.
*   **[Roadmap](ROADMAP.md):** Future plans and development direction.

*(Additional Tutorials and How-To Guides may be added to the `docs/` directory.)*

## Important: Ignoring Generated Files

It is strongly recommended to add generated files like `contextvibes.md`, `contextvibes_ai_trace.log`, `*.log`, and `tfplan.out` to your project's `.gitignore` file.

## Terminal Output vs. AI Log File

Context Vibes uses two distinct output mechanisms:

1.  **Terminal Output (stdout/stderr):** For human readability and high-level status/errors. Uses structured prefixes (`SUMMARY:`, `INFO:`, etc.).
2.  **AI Log File (JSON):** Written to `contextvibes_ai_trace.log` by default (configurable). Contains a detailed, structured trace for AI analysis or debugging.

## Code of Conduct

Act professionally and respectfully. Be kind, considerate, and welcoming. Harassment or exclusionary behavior will not be tolerated.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.