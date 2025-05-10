## Command Reference (Reference)

This section provides a detailed reference for each command in the Context Vibes CLI.

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
contextvibes diff
```

**Description:**

Generates a Markdown summary of pending Git changes (staged, unstaged, untracked) in the Git repository and **overwrites** the context file: `contextvibes.md`.

**Flags:**

This command has no flags.

**Example Usage:**

```bash
contextvibes diff
```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                   |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The diff summary was generated successfully, or no changes were found.                                                                                   |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: Git command failures. |

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

**Global Flags:**
  - The global `--yes` (or `-y`) flag bypasses confirmation prompts for the daily Git kickoff actions. It does not bypass the interactive setup questions for the strategic kickoff prompt generation.

**Flags:**

| Flag                        | Short | Description                                                                                                | Data Type | Default Value |
|-----------------------------|-------|------------------------------------------------------------------------------------------------------------|-----------|---------------|
| `--branch`                  | `-b`  | Name for the new daily/feature branch (e.g., `feature/JIRA-123-task-name`). Used by daily kickoff mode.    | string    | ""            |
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

### `sync`

**Synopsis:**

```contextvibes sync
```

**Description:**

Syncs the local branch with the remote, ensuring it's clean, pulling with rebase, and pushing if ahead.

**Flags:**

This command has no flags.

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

### `wrapup`

**Synopsis:**

```
contextvibes wrapup
```

**Description:**

Finalizes daily work: stages all changes, commits (with a default message if needed), and pushes the current branch.

**Flags:**

This command has no flags.

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

### `plan`

**Synopsis:**

```
contextvibes plan
```

**Description:**

Generates an execution plan (e.g., terraform plan, pulumi preview) based on the detected project type.

**Flags:**

This command has no flags.

**Example Usage:**

*   Generate a plan:

    ```bash
    contextvibes plan
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                               |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success (no changes detected).  Terraform: no changes detected. Pulumi: preview completed successfully.                            |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details. Common causes: missing tools, invalid configuration files.   |
| 2         | Terraform only: plan indicates changes are needed (success for plan command itself).                                                   |

### `deploy`

**Synopsis:**

```
contextvibes deploy
```

**Description:**

Deploys infrastructure changes (terraform apply, pulumi up) based on the detected project type.

**Flags:**

This command has no flags.

**Example Usage:**

*   Deploy changes:

    ```bash
    contextvibes deploy
    ```

* Deploy changes automatically:

    ```bash
    contextvibes deploy -y
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                  |
|-----------|--------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. Deployment completed successfully.                                                                            |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Missing tfplan.out.  |

### `quality`

**Synopsis:**

```contextvibes quality
```

**Description:**

Runs code formatting and linting checks (Terraform, Python, Go) based on the detected project type.

**Flags:**

This command has no flags.

**Example Usage:**

*   Run quality checks:

    ```bash
    contextvibes quality
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                               |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. All quality checks passed successfully, or warnings were reported but no critical errors occurred.                                         |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. Common causes:  formatting violations, linter errors.  |

### `format`

**Synopsis:**

```
contextvibes format
```

**Description:**

Applies code formatting (go fmt, terraform fmt, isort, black) based on the detected project type.

**Flags:**

This command has no flags.

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

### `test`

**Synopsis:**

```
contextvibes test [args...]
```

**Description:**

Runs project-specific tests (e.g., go test, pytest) based on the detected project type. Any arguments are passed to the test runner.

**Flags:**

This command accepts arbitrary arguments passed to the underlying test runner.

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
| 0         | Success. All tests passed successfully.                                                     |
| 1         | An error occurred. Check the error messages in the terminal output and the AI log file for details. |

### `version`

**Synopsis:**

```
contextvibes version
```

**Description:**

Displays the version number of the Context Vibes CLI.

**Flags:**

This command has no flags.

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