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
contextvibes kickoff [--branch <branch-name>]
```

**Description:**

Performs the start-of-work workflow: updates the main branch from the remote, creates and switches to a new branch, and pushes the new branch to the remote. Requires a clean state on the main branch.

**Flags:**

| Flag       | Short | Description                                                                   | Data Type | Default Value | Overrides Config File |
|------------|-------|-------------------------------------------------------------------------------|-----------|---------------|-----------------------|
| `--branch` | `-b`  | Name for the new branch (e.g., `feature/JIRA-123-task-name`). If omitted, user is prompted. | string    | ""            | No                    |

**Example Usage:**

*   Start a new feature branch with a specified name:

    ```bash
    contextvibes kickoff --branch feature/JIRA-123-new-widget
    ```

*   Start a new fix branch with a specified name, bypassing confirmation:

    ```bash
    contextvibes kickoff -b fix/login-bug -y
    ```

*   Start a new branch, prompting for the name:

    ```bash
    contextvibes kickoff
    ```

**Exit Codes:**

| Exit Code | Meaning                                                                                                                                                                                                               |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0         | Success. The new branch was created and pushed successfully.                                                                                                                                                        |
| 1         | An error occurred. Check the error message in the terminal output and the AI log file for details.  Common causes: dirty working directory, not on main branch, invalid branch name, remote access issues, etc. |

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

```
contextvibes sync
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

```
contextvibes quality
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