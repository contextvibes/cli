---
title: "ContextVibes CLI: Configuration Reference (.contextvibes.yaml)"
artifactVersion: "1.0.0"
summary: "The definitive guide to configuring the ContextVibes CLI using the .contextvibes.yaml file. Details all sections including git, logging, validation, projectState, and ai, explaining how to customize CLI behavior and manage project workflow state."
owner: "Scribe"
createdDate: "2025-06-08T11:20:00Z"
lastModifiedDate: "2025-06-08T11:20:00Z"
defaultTargetPath: "docs/CONFIGURATION_REFERENCE.md"
usageGuidance:
  - "Primary reference for customizing `contextvibes` CLI behavior."
  - "Use to understand how to configure Git settings, logging outputs, and validation rules for branches and commits."
  - "Explains the purpose of the machine-managed `projectState` and `ai.collaborationPreferences` sections."
  - "Consult to learn about the precedence of settings (flags vs. config file vs. defaults)."
tags:
  - "configuration"
  - "reference"
  - "contextvibes"
  - "cli"
  - "yaml"
  - "git-config"
  - "validation-rules"
  - "project-state"
  - "ai-preferences"
  - "nix"
---

# Configuration Reference: `.contextvibes.yaml`

The ContextVibes CLI can be configured using a `.contextvibes.yaml` file located in the root directory of your project. This file allows you to customize various aspects of the CLI's behavior. If the file is not present, or if specific settings are omitted, the CLI uses sensible built-in defaults.

### File Format

The `.contextvibes.yaml` file uses standard YAML syntax.

### The Role of `.contextvibes.yaml`

This configuration file is the central source of truth for the `contextvibes` CLI's behavior and state within your project. It serves two primary functions:

1.  **Behavioral Configuration:** You can define project-specific standards, such as Git branch naming conventions or commit message validation patterns. The CLI reads these settings to enforce your team's workflow. This separates project-specific *policy* from the CLI's core *logic*.
2.  **State Management:** The CLI writes to this file to record the state of certain workflows. For example, after you complete the strategic kickoff, the `projectState` section is updated. This allows the CLI to have a persistent "memory" of key project milestones.

This file is complementary to environment definitions like `.idx/dev.nix`, which installs the necessary tools (`go`, `git`), while `.contextvibes.yaml` configures how `contextvibes` uses those tools.

---

### Top-Level Sections

The configuration file is currently organized into the following top-level sections:

*   `git`: Settings related to Git repository interaction.
*   `logging`: Settings related to logging.
*   `validation`: Settings related to input validation rules.
*   `projectState`: State information managed by `contextvibes` about the project.
*   `ai`: Settings related to AI interaction preferences.

### Section Details

#### `git`

This section configures Git-related settings.

| Key                 | Data Type | Description                                                              | Default Value (Built-in) |
| ------------------- | --------- | ------------------------------------------------------------------------ | ------------------------ |
| `defaultRemote`     | string    | The name of the default Git remote (e.g., for `sync`, `kickoff` push).   | `origin`                 |
| `defaultMainBranch` | string    | The name of the default main branch (e.g., used by `kickoff` as the base). | `main`                   |

**Example:**

```yaml
git:
  defaultRemote: "origin"
  defaultMainBranch: "main"
```

#### `logging`

This section configures logging settings for the AI trace log.

| Key              | Data Type | Description                                                                                                                     | Default Value (Built-in)    |
| ---------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------- | --------------------------- |
| `defaultAILogFile` | string    | The file path for the detailed AI JSON trace log. This setting is overridden by the `--ai-log-file` command-line flag, if provided. | `contextvibes_ai_trace.log` |

**Example:**

```yaml
logging:
  defaultAILogFile: "logs/contextvibes_ai_activity.jsonl" # Path relative to project root
```

#### `validation`

This section configures input validation rules used by various commands.

##### `validation.branchName`

Settings for validating branch names, typically used by the `kickoff` command.

| Key       | Data Type | Description                                                                                                                                                    | Default Value (Built-in)                 |
| --------- | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| `enable`  | boolean   | Whether branch name validation is enabled. If omitted or `null`, defaults to `true`. Set to `false` to disable.                                                 | `true`                                   |
| `pattern` | string    | A Go-compatible regular expression used to validate branch names if `enable` is `true`. If `enable` is `true` and `pattern` is empty, a default is used. | `^((feature|fix|docs|format)/.+)$`       |

**Example:**

```yaml
validation:
  branchName:
    enable: true
    # pattern: "^(feature|fix|chore|task)/[A-Z]+-[0-9]+-.*$" # Example custom pattern
```

##### `validation.commitMessage`

Settings for validating commit messages, used by the `commit` command.

| Key       | Data Type | Description                                                                                                                                                               | Default Value (Built-in)                                                                                                                                                           |
| --------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enable`  | boolean   | Whether commit message validation is enabled. If omitted or `null`, defaults to `true`. Set to `false` to disable.                                                         | `true`                                                                                                                                                                             |
| `pattern` | string    | A Go-compatible regular expression used to validate commit messages if `enable` is `true`. If `enable` is `true` and `pattern` is empty, a default Conventional Commits pattern is used. | `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+$` |

**Example:**

```yaml
validation:
  # ... (branchName settings)
  commitMessage:
    enable: true
    # pattern: "^(TASK-[0-9]+|MERGE): .+" # Example custom pattern
```

#### `run`

This section configures the behavior of the `contextvibes run` command, allowing you to define prerequisite verification checks for example applications.

##### `run.examples`

This is a map where each key is the path to an example directory (e.g., `examples/hello-world`) and the value contains settings for that specific example.

| Key                 | Data Type | Description                                                              |
| ------------------- | --------- | ------------------------------------------------------------------------ |
| `verify`            | array     | A list of verification checks to run before the example is executed.     |

**`verify` object fields:**

| Key           | Data Type | Description                                                      | Required |
|---------------|-----------|------------------------------------------------------------------|----------|
| `name`        | string    | A short, unique name for the check.                              | Yes      |
| `description` | string    | A user-friendly description of what the check does.              | No       |
| `command`     | string    | The command to execute for verification.                         | Yes      |
| `args`        | array     | A list of string arguments to pass to the command.               | No       |

**Example:**

```yaml
run:
  examples:
    "examples/hello-world":
      verify:
        - name: "check-go-version"
          description: "Ensuring Go 1.24+ is installed."
          command: "go"
          args: ["version"]
        - name: "check-for-gh-cli"
          description: "Check if the GitHub CLI is available."
          command: "gh"
          args: ["--version"]
```

#### `projectState`

This section stores state information about the project that is managed by ContextVibes CLI commands. Users should generally not edit this section manually unless specifically instructed.

| Key                         | Data Type | Description                                                                                                                                  | Default Value (Built-in) | Written By Command                               |
| --------------------------- | --------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ | ------------------------------------------------ |
| `strategicKickoffCompleted` | boolean   | Indicates if the comprehensive strategic project kickoff (facilitated by `kickoff --strategic` prompt generation) has been marked as complete. | `false`                  | `contextvibes kickoff --mark-strategic-complete` |
| `lastStrategicKickoffDate`  | string    | An RFC3339 timestamp indicating when the strategic kickoff was last marked as complete. Optional.                                             | `""` (empty string)      | `contextvibes kickoff --mark-strategic-complete` |

**Example (as written by `contextvibes`):**

```yaml
projectState:
  strategicKickoffCompleted: true
  lastStrategicKickoffDate: "2025-05-10T12:00:00Z"
```

#### `ai`

This section configures preferences related to AI interaction, specifically for how the ContextVibes CLI itself should behave during setup phases or if it directly generates AI-assisted content in the future.

##### `ai.collaborationPreferences`

These preferences are typically set or confirmed during the initial interactive phase of `contextvibes kickoff --strategic` and are used to tailor how the CLI (or the generated master prompt for an external AI) suggests interactions.

| Key                     | Data Type | Description                                                                                                | Default Value (Built-in)              |
| ----------------------- | --------- | ---------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| `codeProvisioningStyle` | string    | How the CLI should offer code snippets. Options: `bash_cat_eof`, `raw_markdown`.                           | `bash_cat_eof`                        |
| `markdownDocsStyle`     | string    | How the CLI should offer Markdown documentation. Options: `raw_markdown`.                                  | `raw_markdown`                        |
| `detailedTaskMode`      | string    | Preferred interaction mode for detailed tasks. Options: `mode_a` (Generate & Refine), `mode_b` (Interactive). | `mode_b`                              |
| `proactiveDetailLevel`  | string    | Level of detail in AI explanations. Options: `detailed_explanations`, `concise_unless_asked`.                 | `detailed_explanations`               |
| `aiProactivity`         | string    | How proactive the AI should be in offering suggestions. Options: `proactive_suggestions`, `wait_for_request`. | `proactive_suggestions`               |

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

---

### Precedence

The configuration settings are applied in the following order of precedence (highest to lowest):

1.  **Command-line flags:** Flags provided directly when running a command (e.g., `--ai-log-file`, `--log-level-ai`, global `--yes`) always override any other settings.
2.  **`.contextvibes.yaml` file:** Settings defined in this file in the project root override the built-in defaults if the file exists and the setting is specified.
3.  **Built-in Defaults:** The default values hardcoded within the CLI application (defined in `internal/config/config.go`).

This means that if a setting is specified both in the configuration file and as a command-line flag, the command-line flag will take precedence. If no config file is found, or the setting isn't specified in the config file or via a flag, the built-in default value will be used.
```
