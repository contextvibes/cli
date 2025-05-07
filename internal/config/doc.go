/*
Package config manages the configuration for the contextvibes CLI application.

It defines the structure of the configuration, provides functions to load
configuration from a YAML file (defaulting to '.contextvibes.yaml' in the
project root), and offers a way to get default configuration values.

The primary components are:
  - Config: The main struct holding all configuration settings, including Git behavior,
    logging preferences, and validation rules for branch names and commit messages.
  - GitSettings, LoggingSettings, ValidationRule: Sub-structs organizing related
    configuration items.
  - GetDefaultConfig(): Returns a pointer to a Config struct populated with sensible
    default values.
  - LoadConfig(filePath string): Attempts to load configuration from the specified
    YAML file. Returns nil if the file doesn't exist, allowing graceful fallback to defaults.
  - FindRepoRootConfigPath(execClient *exec.ExecutorClient): Locates the configuration
    file by searching upwards from the current directory to the Git repository root.
  - MergeWithDefaults(loadedCfg *Config, defaultConfig *Config): Merges a loaded
    user configuration with the default configuration, giving precedence to user-defined values.

Constants are also defined for default filenames (e.g., DefaultConfigFileName,
DefaultCodemodFilename, DefaultDescribeOutputFile, UltimateDefaultAILogFilename)
and default patterns for validation. These constants are intended to be used by
CLI commands to ensure consistent default behavior.

The typical flow involves:
1. Attempting to find and load a user-defined '.contextvibes.yaml' file.
2. If found and valid, merging it with the application's default configuration.
3. If not found or invalid, using the application's default configuration directly.
The resulting configuration is then used throughout the application, particularly
by the cmd package to influence command behavior.
*/
package config
