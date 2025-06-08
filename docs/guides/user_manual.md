---
title: "ContextVibes CLI: Comprehensive User Manual"
artifactVersion: "1.0.0"
summary: "The definitive user manual for the ContextVibes CLI. Covers installation, configuration, all commands with examples, core concepts like the dual output system, and integration with AI workflows."
owner: "Scribe"
createdDate: "2025-06-08T00:00:00Z"
lastModifiedDate: "2025-06-08T10:30:00Z" # I've updated this to reflect our changes
defaultTargetPath: "manuals/contextvibes_cli_manual.md"
usageGuidance:
  - "Primary reference for all `contextvibes` CLI command syntax, flags, and examples."
  - "Use to understand how to install and configure the CLI using `.contextvibes.yaml`."
  - "Consult for explanations of core concepts like the strategic kickoff workflow and AI context generation."
  - "Provides guidance on contributing to the CLI and setting up a local development environment."
tags:
  # Core Identifiers
  - "contextvibes"
  - "cli"
  - "user-manual"
  - "documentation"
  # Key Features & Commands
  - "command-reference"
  - "installation"
  - "configuration"
  - "kickoff-command"
  - "describe-command"
  - "commit-command"
  - "sync-command"
  - "git-workflow"
  # Concepts
  - "ai-context"
  - "thea-framework"
  - "slog"
  - "cobra"
---
# ContextVibes CLI User Manual

## Introduction

Welcome to the ContextVibes CLI User Manual. This document provides a comprehensive reference for all commands, flags, and functionalities of the ContextVibes Command Line Interface.

ContextVibes CLI is designed as a developer co-pilot to streamline common development tasks, enhance productivity, and generate rich context for AI-assisted software engineering. It wraps common tools and workflows, providing consistent interfaces, clear terminal output, and detailed background logging suitable for AI consumption.

This manual is organized as follows:
*   **General Usage:** Basic patterns for interacting with the CLI.
*   **Global Flags:** Flags applicable to most or all commands.
*   **Command Reference:** Detailed descriptions of each command, listed alphabetically.

We encourage you to familiarize yourself with the commands relevant to your workflow. For project-specific configuration options, please refer to the `CONFIGURATION_REFERENCE.md`.

## General Usage

The ContextVibes CLI follows standard command-line patterns.

*   **Executing Commands:**
    ```bash
    contextvibes [command] [subcommand] [arguments] [flags]
    ```
*   **Getting Help:**
    *   For a list of all commands:
        ```bash
        contextvibes --help
        ```
    *   For help on a specific command:
        ```bash
        contextvibes [command] --help
        ```
    *   For help on a specific subcommand:
        ```bash
        contextvibes [command] [subcommand] --help
        ```

## Global Flags

These flags can be used with any command:

| Flag             | Short | Description                                                                                                                                  | Data Type | Default Value                  | Overrides Config File |
|------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------|-----------|--------------------------------|-----------------------|
| `--yes`          | `-y`  | Assume 'yes' to all confirmation prompts, enabling non-interactive mode.                                                                       | boolean   | `false`                        | No                    |
| `--ai-log-file`  |       | Path for the detailed AI JSON trace log.                                                                                                       | string    | From config or `contextvibes_ai_trace.log` | Yes                   |
| `--log-level-ai` |       | Minimum level for the AI log file (debug, info, warn, error).                                                                                  | string    | `debug`                        | Yes                   |

---

## Command Reference

This section provides a detailed reference for each command available in the ContextVibes CLI, listed alphabetically.

### `codemod`

**Synopsis:**

```
contextvibes codemod [--script <file.json>]
```

**Description:**

Applies programmatic code modifications or deletions from a JSON script.

**Flags:**

| Flag       | Short | Description                                              | Data Type | Default Value | Overrides Config File |
|------------|-------|----------------------------------------------------------|-----------|---------------|-----------------------|
| `--script` | `-s`  | Path to the JSON codemod script file.                  | string    | `codemod.json`| No                    |

**Example Usage:**

*   Run the codemod using the default script:

    ```bash
    contextvibes codemod
    ```

*   Run the codemod using a custom script:

    ```bash
    contextvibes codemod --script ./my_refactor_script.json
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                       |
|-----------|-----------------------------------------------------------------------------------------------|
| 0         | Success. The codemod script was executed successfully.                                         |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. |

### `commit`

**Synopsis:**

```
contextvibes commit -m <message>
```

**Description:**

Stages all changes and commits locally with a provided message.  Commit message validation is active by default, and the rules are configurable via `.contextvibes.yaml`.

**Flags:**

| Flag        | Short | Description                         | Data Type | Default Value | Overrides Config File |
|-------------|-------|-------------------------------------|-----------|---------------|-----------------------|
| `--message` | `-m`  | Commit message (required).          | string    | ""            | No                    |

**Example Usage:**

*   Commit changes with a message:

    ```bash
    contextvibes commit -m "feat(auth): Implement OTP login"
    ```

*   Commit changes with a message, bypassing confirmation:

    ```bash
    contextvibes commit -m "fix(api): Correct typo in user model" -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                                                                     |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The changes were staged and committed successfully.                                                                                                                                               |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: missing commit message, invalid commit message format, Git command failures, etc. |

### `deploy`

**Synopsis:**

```
contextvibes deploy
```

**Description:**

Deploys infrastructure changes (terraform apply, pulumi up) based on the detected project type.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

*   Deploy changes:

    ```bash
    contextvibes deploy
    ```

* Deploy changes automatically (using global flag):

    ```bash
    contextvibes deploy -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                  |
|-----------|--------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. Deployment completed successfully.                                                                            |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: Missing `tfplan.out` for Terraform, tool execution errors.  |

### `describe`

**Synopsis:**

```
contextvibes describe [-o <output_file>]
```

**Description:**

Gathers project context (user prompt, environment, git status, structure, relevant files) and writes it to a Markdown file, suitable for AI interaction. The default output file is `contextvibes.md`.

**Flags:**

| Flag          | Short | Description                                                              | Data Type | Default Value      | Overrides Config File |
|---------------|-------|--------------------------------------------------------------------------|-----------|--------------------|-----------------------|
| `--output`    | `-o`  | Path to write the context markdown file.                                 | string    | `contextvibes.md`  | No                    |

**Example Usage:**

*   Generate a context file with the default name:

    ```bash
    contextvibes describe
    ```

*   Generate a context file with a custom name:

    ```bash
    contextvibes describe -o project_snapshot.md
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                   |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The context file was generated successfully.                                                                                   |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: empty prompt, read failures. |

### `diff`

**Synopsis:**

```
contextvibes diff```

**Description:**

Generates a Markdown summary of pending Git changes (staged, unstaged, untracked) in the Git repository and **overwrites** the context file: `contextvibes.md`.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

```bash
contextvibes diff
```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                   |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The diff summary was generated successfully, or no changes were found.                                                                                   |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: Git command failures. |

### `format`

**Synopsis:**

```
contextvibes format
```

**Description:**

Applies code formatting (go fmt, terraform fmt, isort, black) based on the detected project type.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

*   Apply code formatting:

    ```bash
    contextvibes format
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                       |
|-----------|-----------------------------------------------------------------------------------------------|
| 0         | Success. All formatting tools completed successfully or applied changes.                       |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. |

### `index`

**Synopsis:**

```
contextvibes index [--thea-path <path>] [--template-path <path>] [-o <output-file>]
```

**Description:**

Indexes documents (e.g., THEA framework files, project templates) from specified local directories to create a structured JSON manifest. It parses YAML front matter from Markdown files to extract metadata like title, version, summary, usage guidance, owner, dates, and tags. The `id` of each document is derived from its relative file path.

This manifest is intended for consumption by other tools or AI models to understand available project artifacts.

**Flags:**

| Flag              | Short | Description                                                              | Data Type | Default Value            | Overrides Config File |
|-------------------|-------|--------------------------------------------------------------------------|-----------|--------------------------|-----------------------|
| `--thea-path`     |       | Path to the root of a THEA-like structured directory to index.           | string    | `""`                     | No                    |
| `--template-path` |       | Path to the root of a project template directory to index.               | string    | `""`                     | No                    |
| `--output`        | `-o`  | Output path for the generated JSON manifest file.                        | string    | `project_manifest.json`  | No                    |

**Example Usage:**

*   Index THEA documents and project templates, saving to a custom file:
    ```bash
    contextvibes index --thea-path ../THEA-main/docs --template-path ../THEA-main/templates -o my_project_manifest.json
    ```
*   Index only THEA documents with the default output file name:
    ```bash
    contextvibes index --thea-path /path/to/thea_docs
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                   |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The document manifest was generated successfully. This includes cases where no documents were found (resulting in an empty manifest). |
| 1         | An error occurred. Common causes: invalid paths, read permissions issues, errors parsing front matter, or errors writing the output file. Check terminal output and the AI log file. |

### `kickoff`

**Synopsis:**

```
contextvibes kickoff [--branch <branch-name>] [--strategic] [--mark-strategic-complete]
```

**Description:**

Manages project kickoff workflows.

**Default Behavior (Daily Kickoff, if strategic completed):**
  - Requires a clean state on the main branch (configurable, default: `main`).
  - Updates the main branch from the remote (configurable, default: `origin`).
  - Creates and switches to a new daily/feature branch. The name is taken from the `--branch` flag or prompted for if the flag is omitted (and `--yes` is not active).
  - Branch name validation is applied based on rules in `.contextvibes.yaml` (default pattern: `^((feature|fix|docs|format)/.+)$`).
  - Pushes the new branch to the remote and sets upstream tracking.

**Strategic Kickoff Prompt Generation Mode (`--strategic`, or if first run in a project):**
  - This mode is triggered if `projectState.strategicKickoffCompleted` is `false` (or not set) in `.contextvibes.yaml`, or if the `--strategic` flag is explicitly used.
  - Conducts a brief interactive session to gather:
    1.  User preferences for how ContextVibes CLI should format its own outputs (e.g., code blocks, Markdown style) and interact during setup. These are saved to `.contextvibes.yaml` under `ai.collaborationPreferences`.
    2.  Basic project details (name, primary application type, current stage like new/existing).
    3.  Confirmation of ContextVibes CLI readiness (e.g., installed, ENV vars).
  - Generates a comprehensive master prompt file named `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md` in the project root.
  - This generated file contains a detailed protocol (based on the embedded template `internal/kickoff/assets/strategic_kickoff_protocol_template.md`), parameterized with the gathered project details and CLI preferences.
  - The user is then instructed to take this `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md` file and use its content as the initial prompt for their chosen external AI assistant (e.g., Gemini, Claude, ChatGPT). The external AI will then facilitate the detailed strategic kickoff discussion.
  - The generated master prompt also instructs the external AI to ask the user to run other `contextvibes` commands (like `describe`, `status`) as needed during their session to provide live project data back to the AI. It also instructs the AI to generate structured YAML for certain user decisions (like the `ai.collaborationPreferences`).

**Marking Strategic Kickoff as Complete (`--mark-strategic-complete`):**
  - This flag is used *after* the user has completed their AI-guided strategic kickoff session.
  - It updates the project's `.contextvibes.yaml` file by:
    - Setting `projectState.strategicKickoffCompleted: true`.
    - Recording `projectState.lastStrategicKickoffDate` with the current timestamp.
    - Persisting any `ai.collaborationPreferences` that were gathered during the most recent `--strategic` run's setup phase.
  - This enables the daily Git kickoff workflow for subsequent `contextvibes kickoff` runs (without `--strategic`).

**Global Flags Interaction:**
  - The global `--yes` (or `-y`) flag bypasses confirmation prompts for the daily Git kickoff actions. It does not bypass the interactive setup questions for the strategic kickoff prompt generation.

**Flags:**

| Flag                        | Short | Description                                                                                                | Data Type | Default Value |
|-----------------------------|-------|------------------------------------------------------------------------------------------------------------|-----------|---------------|
| `--branch`                  | `-b`  | Name for the new daily/feature branch (e.g., `feature/JIRA-123-task-name`). Used by daily kickoff mode.    | string    | `""`          |
| `--strategic`               |       | Forces the strategic kickoff prompt generation, even if a previous strategic kickoff was marked complete.    | boolean   | `false`       |
| `--mark-strategic-complete` |       | Marks the strategic kickoff as complete in `.contextvibes.yaml`. Mutually exclusive with other flags.      | boolean   | `false`       |


**Example Usage:**

*   **Daily Kickoff (assuming strategic kickoff is marked complete):**
    ```bash
    contextvibes kickoff --branch feature/JIRA-123-new-widget
    contextvibes kickoff -b fix/login-bug -y
    contextvibes kickoff # Prompts for branch name if not provided and not -y
    ```

*   **Strategic Kickoff Prompt Generation:**
    ```bash
    contextvibes kickoff --strategic # Always runs the strategic prompt generation
    contextvibes kickoff             # Runs strategic prompt generation if first time in project
    ```
    *(Then, take `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md` to your AI)*

*   **Mark Strategic Kickoff as Done:**
    ```bash
    contextvibes kickoff --mark-strategic-complete
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                                                                               |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The requested kickoff operation (daily branch creation, strategic prompt generation, or marking complete) was successful.                                                                                 |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: Git prerequisites not met (dirty WD, wrong branch), invalid branch name, file I/O errors, etc. |

### `plan`

**Synopsis:**

```
contextvibes plan
```

**Description:**

Generates an execution plan (e.g., terraform plan, pulumi preview) based on the detected project type.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

*   Generate a plan:

    ```bash
    contextvibes plan
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                               |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. Terraform: plan indicates no changes. Pulumi: preview completed successfully (may or may not have changes).                 |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: missing tools, invalid configuration files, tool execution errors. |
| 2         | Terraform only: plan indicates changes are needed (considered a successful outcome for the plan command itself).                     |

### `quality`

**Synopsis:**

```contextvibes quality
```

**Description:**

Runs code formatting and linting checks (Terraform, Python, Go) based on the detected project type.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

*   Run quality checks:

    ```bash
    contextvibes quality
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                               |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. All critical quality checks passed successfully. Warnings may have been reported for non-critical issues (e.g., some linters).             |
| 1         | An error occurred or critical checks failed. Check the error messages in the terminal output and the AI log file for details. Common causes:  formatting violations, linter errors, validation failures.  |

### `status`

**Synopsis:**

```contextvibes status
```

**Description:**

Shows a concise summary of the working tree status using `git status --short`. This includes staged changes, unstaged changes, and untracked files.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

```bash
contextvibes status
```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                   |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The Git status was displayed successfully.                                                                                       |
| 1         | An error occurred. Common causes: not a Git repository, `git` command execution failure. Check terminal output and AI log.                 |

### `sync`

**Synopsis:**

```contextvibes sync
```

**Description:**

Syncs the local branch with the remote, ensuring it's clean, pulling with rebase, and pushing if ahead.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

*   Sync the current branch:

    ```bash
    contextvibes sync
    ```

*   Sync the current branch, bypassing confirmation:

    ```bash
    contextvibes sync -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                           |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The branch was synced successfully.                                                                                                                    |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: dirty working directory, pull rebase failures, remote access issues.  |

### `test`

**Synopsis:**

```
contextvibes test [args...]
```

**Description:**

Runs project-specific tests (e.g., go test, pytest) based on the detected project type. Any arguments are passed to the test runner.

**Flags:**

This command accepts arbitrary arguments which are passed directly to the underlying test runner.

**Example Usage:**

*   Run tests:

    ```bash
    contextvibes test
    ```

* Run verbose go tests:

    ```bash
    contextvibes test -v
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                     |
|-----------|---------------------------------------------------------------------------------------------|
| 0         | Success. All tests passed successfully, or no tests were executed for the project type.     |
| 1         | An error occurred or tests failed. Check the error messages in the terminal output and the AI log file for details. |

### `thea`

This command serves as a parent for interacting with THEA framework artifacts and services.

**Synopsis:**

```
contextvibes thea <subcommand> [subcommand-flags]
```

**Description:**

Provides subcommands to fetch artifacts from the THEA framework repository, and potentially other interactions in the future.

**Flags:**

This parent command has no specific flags other than global flags. Flags are specific to its subcommands.

**Subcommands:**

*   `get-artifact`: Fetches a specific artifact.
*   `index`: (Note: This was a root command `contextvibes index`. If it's meant to be under `thea` like `contextvibes thea index`, that's a CLI structural change. Current docs reflect `contextvibes index`.)

**Example Usage:**

*   Get a specific THEA artifact:
    ```bash
    contextvibes thea get-artifact playbooks/project_initiation/master_strategic_kickoff_prompt -o kickoff_template.md
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                    |
|-----------|----------------------------------------------------------------------------|
| 0         | Success (if a valid subcommand was executed successfully).                  |
| 1         | An error occurred, such as an invalid subcommand or an error within a subcommand. |

---

#### `thea get-artifact`

**Synopsis:**

```contextvibes thea get-artifact <artifact-id> [--version <version>] [--output <file>] [--force]
```

**Argument:**

*   `<artifact-id>`: (Required) The unique ID of the THEA artifact to fetch (e.g., `playbooks/project_initiation/master_strategic_kickoff_prompt`).

**Description:**

Downloads a specified artifact (e.g., playbook, template, guide) from the central THEA framework repository using its unique artifact ID. The artifact ID typically follows a path-like structure. The fetched content is saved to a local file. Default THEA repository URLs are used.

**Flags:**

| Flag        | Short | Description                                                                                                | Data Type | Default Value | Overrides Config File |
|-------------|-------|------------------------------------------------------------------------------------------------------------|-----------|---------------|-----------------------|
| `--version` | `-v`  | Version hint (e.g., git tag/branch like `v0.7.0` or `main`) for the artifact.                                | string    | `""`          | No                    |
| `--output`  | `-o`  | Path to save the fetched artifact. If empty, uses a default name derived from artifact metadata or ID.     | string    | `""`          | No                    |
| `--force`   | `-f`  | Overwrite the output file if it already exists.                                                              | boolean   | `false`       | No                    |

**Example Usage:**

*   Fetch an artifact and save it to a specified file:
    ```bash
    contextvibes thea get-artifact playbooks/project_initiation/master_strategic_kickoff_prompt -o kickoff_prompt.md
    ```
*   Fetch an artifact using a version hint (e.g., a specific tag):
    ```bash
    contextvibes thea get-artifact docs/style-guide --version v1.2.0
    ```
*   Fetch an artifact and overwrite an existing local file:
    ```bash
    contextvibes thea get-artifact playbooks/common/README -o COMMON_README.md --force
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                           |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The artifact was fetched and saved successfully.                                                                                                        |
| 1         | An error occurred. Common causes: artifact ID not found in manifest, network error fetching manifest or content, file system error writing output. Check terminal output and AI log. |

### `version`

**Synopsis:**

```
contextvibes version
```

**Description:**

Displays the version number of the Context Vibes CLI.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

*   Display the version:

    ```bash
    contextvibes version
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                    |
|-----------|----------------------------------------------------------------------------|
| 0         | Success. The version number was displayed successfully.                   |
| 1         | An error occurred. Check the terminal and AI log file, though this is unlikely.  |

### `wrapup`

**Synopsis:**

```
contextvibes wrapup
```

**Description:**

Finalizes daily work: stages all changes, commits (with a default message if needed), and pushes the current branch.

**Flags:**

This command has no specific flags other than global flags.

**Example Usage:**

*   Wrap up the current branch:

    ```bash
    contextvibes wrapup
    ```

*   Wrap up the current branch, bypassing confirmation:

    ```bash
    contextvibes wrapup -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                              |
|-----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The changes were staged, committed (if needed), and the branch was pushed (if needed).                                                                     |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: git failures, merge conflicts, or remote errors. |