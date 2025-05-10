## Configuration Reference (Reference)

The ContextVibes CLI can be configured using a `.contextvibes.yaml` file located in the root directory of your project. This file allows you to customize various aspects of the CLI's behavior. If the file is not present, or if specific settings are omitted, the CLI uses sensible built-in defaults.

### File Format

The `.contextvibes.yaml` file uses YAML syntax.

### Top-Level Sections

The configuration file is currently organized into the following top-level sections:

*   `git`: Settings related to Git repository interaction.
*   `logging`: Settings related to logging.
*   `validation`: Settings related to input validation rules.
*   `projectState`: State information managed by ContextVes about the project.
*   `ai`: Settings related to AI interaction preferences.

### Section Details

#### `git`

This section configures Git-related settings.

| Key                  | Data Type | Description                                                                                                                      | Default Value (Built-in) |
|----------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------|--------------------------|
| `defaultRemote`      | string    | The name of the default Git remote (e.g., for `sync`, `kickoff` push).                                                             | `origin`                 |
| `defaultMainBranch`  | string    | The name of the default main branch (e.g., used by `kickoff` as the base).                                                      | `main`                   |

**Example:**
```yaml
git:
  defaultRemote: "origin"
  defaultMainBranch: "main"
```

#### `logging`

This section configures logging settings for the AI trace log.

| Key               | Data Type | Description                                                                                                                               | Default Value (Built-in)    |
|-------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------|
| `defaultAILogFile` | string    | The file path for the detailed AI JSON trace log. This setting is overridden by the `--ai-log-file` command-line flag, if provided.        | `contextvibes_ai_trace.log` |

**Example:**
```yaml
logging:
  defaultAILogFile: "logs/contextvibes_ai_activity.jsonl" # Path relative to project root
```

#### `validation`

This section configures input validation rules used by various commands.

##### `validation.branchName`

Settings for validating branch names, typically used by the `kickoff` command.

| Key       | Data Type | Description                                                                                                                                          | Default Value (Built-in)                 |
|-----------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------|
| `enable`  | boolean   | Whether branch name validation is enabled. If omitted or `null`, defaults to `true`. Set to `false` to disable.                                         | `true`                                   |
| `pattern` | string    | A Go compatible regular expression pattern used to validate branch names if `enable` is `true`. If `enable` is `true` and `pattern` is empty or not set, a default pattern (`^((feature|fix|docs|format)/.+)$`) is used. | `^((feature|fix|docs|format)/.+)$`       |

**Example:**
```yaml
validation:
  branchName:
    enable: true 
    # pattern: "^(feature|fix|chore|task)/[A-Z]+-[0-9]+-.*$" # Example custom pattern
```

##### `validation.commitMessage`

Settings for validating commit messages, used by the `commit` command.

| Key       | Data Type | Description                                                                                                                                            | Default Value (Built-in)                                                                                                                                                                |
|-----------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `enable`  | boolean   | Whether commit message validation is enabled. If omitted or `null`, defaults to `true`. Set to `false` to disable.                                        | `true`                                                                                                                                                                      |
| `pattern` | string    | A Go compatible regular expression pattern used to validate commit messages if `enable` is `true`. If `enable` is `true` and `pattern` is empty or not set, a default Conventional Commits pattern (`^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+$`) is used. | `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+$` |

**Example:**
```yaml
validation:
  # ... (branchName settings)
  commitMessage:
    enable: true
    # pattern: "^(TASK-[0-9]+|MERGE): .+" # Example custom pattern
```

#### `projectState`

This section stores state information about the project that is managed by ContextVibes CLI commands. Users should generally not edit this section manually unless specifically instructed.

| Key                            | Data Type | Description                                                                                                                                  | Default Value (Built-in) | Written By Command                                  |
|--------------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------------|--------------------------|-----------------------------------------------------|
| `strategicKickoffCompleted`    | boolean   | Indicates if the comprehensive strategic project kickoff (facilitated by `kickoff --strategic` prompt generation) has been marked as complete. | `false`                  | `contextvibes kickoff --mark-strategic-complete`    |
| `lastStrategicKickoffDate`     | string    | An RFC3339 timestamp indicating when the strategic kickoff was last marked as complete. Optional.                                             | `""` (empty string)      | `contextvibes kickoff --mark-strategic-complete`    |

**Example (as written by `contextvibes`):**
```yaml
projectState:
  strategicKickoffCompleted: true
  lastStrategicKickoffDate: "2025-05-10T12:00:00Z"```

#### `ai`

This section configures preferences related to AI interaction, specifically for how the ContextVibes CLI itself should behave during setup phases or if it directly generates AI-assisted content in the future.

##### `ai.collaborationPreferences`

These preferences are typically set or confirmed during the initial interactive phase of `contextvibes kickoff --strategic` and are used to tailor how the CLI (or the generated master prompt for an external AI) suggests interactions.

| Key                       | Data Type | Description                                                                                                | Default Value (Built-in)           |
|---------------------------|-----------|------------------------------------------------------------------------------------------------------------|------------------------------------|
| `codeProvisioningStyle`   | string    | How the CLI (or AI guided by its prompts) should offer code snippets. Options: `bash_cat_eof`, `raw_markdown`. | `bash_cat_eof`                     |
| `markdownDocsStyle`       | string    | How the CLI (or AI) should offer Markdown documentation. Options: `raw_markdown`.                             | `raw_markdown`                     |
| `detailedTaskMode`        | string    | Preferred interaction mode for detailed tasks (e.g., checklists). Options: `mode_a` (Generate & Refine), `mode_b` (Interactive Step-by-Step). | `mode_b`                           |
| `proactiveDetailLevel`    | string    | Level of detail in AI explanations/suggestions. Options: `detailed_explanations`, `concise_unless_asked`.    | `detailed_explanations` (esp. for mode_b) |
| `aiProactivity`           | string    | How proactive the AI should be in offering suggestions. Options: `proactive_suggestions`, `wait_for_request`. | `proactive_suggestions`            |

**Example (as written by `contextvibes` after `kickoff --strategic` setup):**
```yaml
ai:
  collaborationPreferences:
    codeProvisioningStyle: "bash_cat_eof"
    markdownDocsStyle: "raw_markdown"
    detailedTaskMode: "mode_b"
    proactiveDetailLevel: "detailed_explanations"
    aiProactivity: "proactive_suggestions"
```

### Precedence

The configuration settings are applied in the following order of precedence (highest to lowest):

1.  **Command-line flags:** Flags provided directly when running a command (e.g., `--ai-log-file`, `--log-level-ai`, global `--yes`) always override any other settings.
2.  **`.contextvibes.yaml` file:** Settings defined in this file in the project root override the built-in defaults if the file exists and the setting is specified.
3.  **Built-in Defaults:** The default values hardcoded within the CLI application (defined in `internal/config/config.go`).

This means that if a setting is specified both in the configuration file and as a command-line flag, the command-line flag will take precedence. If no config file is found, or the setting isn't specified in the config file or via a flag, the built-in default value will be used.
EOF
```

**Key Changes in this `CONFIGURATION_REFERENCE.md`:**

1.  **Added `projectState` section:** Details `strategicKickoffCompleted` and `lastStrategicKickoffDate`, their types, defaults, and which command writes them.
2.  **Added `ai` section with `collaborationPreferences` sub-section:** Details all the new AI collaboration preference keys, their purpose, example options, and default values. Explains they are set during `kickoff --strategic` setup.
3.  **Updated "Top-Level Sections" list.**
4.  Minor wording adjustments for clarity on defaults (specifying "Built-in" to distinguish from what might end up in a user's file if they accept all defaults during setup).