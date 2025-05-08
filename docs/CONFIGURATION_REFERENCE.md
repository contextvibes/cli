## Configuration Reference (Reference)

The Context Vibes CLI can be configured using a `.contextvibes.yaml` file located in the root directory of your project. This file allows you to customize various aspects of the CLI's behavior. If the file is not present, the CLI uses sensible built-in defaults.

### File Format

The `.contextvibes.yaml` file uses YAML syntax.

### Top-Level Sections

The configuration file is divided into the following top-level sections:

*   `git`: Settings related to Git repository interaction.
*   `logging`: Settings related to logging.
*   `validation`: Settings related to input validation.

### Section Details

#### `git`

This section configures Git-related settings.

| Key                  | Data Type | Description                                                                                                                      | Default Value |
|----------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------|---------------|
| `defaultRemote`      | string    | The name of the default Git remote.                                                                                              | `origin`        |
| `defaultMainBranch`  | string    | The name of the default main branch (used by `kickoff` and other commands).                                                      | `main`          |

Example:

```yaml
git:
  defaultRemote: origin
  defaultMainBranch: main```

#### `logging`

This section configures logging settings.

| Key               | Data Type | Description                                                                                                                               | Default Value             |
|-------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------|---------------------------|
| `defaultAILogFile` | string    | The file path for the detailed AI JSON log. This setting is overridden by the `--ai-log-file` command-line flag, if provided.        | `contextvibes_ai_trace.log` |

Example:

```yaml
logging:
  defaultAILogFile: "logs/contextvibes_ai.log" # Path relative to project root
```

#### `validation`

This section configures input validation rules.

##### `validation.branchName`

| Key       | Data Type | Description                                                                                                                                          | Default Value                                    |
|-----------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------|
| `enable`  | boolean   | Whether branch name validation is enabled. If not set, defaults to `true`.                                                                            | `true`                                           |
| `pattern` | string    | A regular expression pattern used to validate branch names. If `enable` is `true` and `pattern` is not set, a default pattern is used.    | `^((feature|fix|docs|format)/.+)$`               |

##### `validation.commitMessage`

| Key       | Data Type | Description                                                                                                                                            | Default Value                                                                                                                                                                |
|-----------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `enable`  | boolean   | Whether commit message validation is enabled. If not set, defaults to `true`.                                                                            | `true`                                                                                                                                                                      |
| `pattern` | string    | A regular expression pattern used to validate commit messages. If `enable` is `true` and `pattern` is not set, a default Conventional Commits pattern is used. | `^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+$` |

Example:

```yaml
validation:
  branchName:
    enable: true # Default is true. Set to false to disable validation.
    # Default pattern if enabled and not specified: ^((feature|fix|docs|format)/.+)$
    pattern: "^(feature|fix|chore|task)/[A-Z]+-[0-9]+-.*$" # Example custom pattern
  commitMessage:
    enable: true # Default is true. Set to false to disable validation.
    # Default pattern if enabled and not specified: ^(BREAKING|feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-zA-Z0-9\-_/]+\))?:\s.+
    pattern: "^(feat|fix|chore|docs)!?(\\(.+\\))?: .+" # Example custom pattern
```

### Precedence

The configuration settings are applied in the following order of precedence (highest to lowest):

1.  **Command-line flags:** Flags provided directly when running a command (e.g., `--ai-log-file`, `--log-level-ai`) always override any other settings.
2.  **`.contextvibes.yaml` file:** Settings defined in this file override the built-in defaults if the file exists and the setting is specified.
3.  **Built-in Defaults:** The default values hardcoded within the CLI application (defined in `internal/config/config.go`).

This means that if a setting is specified both in the configuration file and as a command-line flag, the command-line flag will take precedence. If no config file is found, or the setting isn't specified in the config file or via a flag, the built-in default value will be used.