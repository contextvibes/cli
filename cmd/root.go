// Package cmd provides the entry point for the ContextVibes CLI.
package cmd

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/contextvibes/cli/cmd/craft"
	"github.com/contextvibes/cli/cmd/factory"
	"github.com/contextvibes/cli/cmd/feedback"
	"github.com/contextvibes/cli/cmd/library"
	"github.com/contextvibes/cli/cmd/product"
	"github.com/contextvibes/cli/cmd/project"
	"github.com/contextvibes/cli/cmd/version"
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/spf13/cobra"
)

const (
	// filePermUserRW represents read/write permissions for the user (0600).
	filePermUserRW = 0o600
)

//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var rootCmd = &cobra.Command{
	Use:   "contextvibes",
	Short: "Manages project tasks: AI context generation, Git workflow, IaC, etc.",
	Long:  `ContextVibes: Your Project Development Assistant CLI.`,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		bootstrapOSExecutor := exec.NewOSCommandExecutor(slog.New(slog.DiscardHandler))
		bootstrapExecClient := exec.NewClient(bootstrapOSExecutor)

		defaultCfg := config.GetDefaultConfig()
		repoConfigPath, _ := config.FindRepoRootConfigPath(bootstrapExecClient)
		if repoConfigPath != "" {
			loadedUserConfig, _ := config.LoadConfig(repoConfigPath)
			if loadedUserConfig != nil {
				globals.LoadedAppConfig = config.MergeWithDefaults(loadedUserConfig, defaultCfg)
			} else {
				globals.LoadedAppConfig = defaultCfg
			}
		} else {
			globals.LoadedAppConfig = defaultCfg
		}

		aiLevel := parseLogLevel(logLevelAIValue, slog.LevelDebug)
		aiOut := io.Discard
		loggingEnabled := (globals.LoadedAppConfig.Logging.Enable != nil && *globals.LoadedAppConfig.Logging.Enable) ||
			aiLogFileFlagValue != ""
		if loggingEnabled {
			targetAILogFile := globals.LoadedAppConfig.Logging.DefaultAILogFile
			if aiLogFileFlagValue != "" {
				targetAILogFile = aiLogFileFlagValue
			}

			logFileHandle, errLogFile := os.OpenFile(
				targetAILogFile,
				os.O_CREATE|os.O_WRONLY|os.O_APPEND,
				filePermUserRW,
			)
			if errLogFile == nil {
				aiOut = logFileHandle
			}
		}
		globals.AppLogger = slog.New(
			slog.NewJSONHandler(aiOut, &slog.HandlerOptions{
				Level:       aiLevel,
				AddSource:   false,
				ReplaceAttr: nil,
			}),
		)

		mainOSExecutor := exec.NewOSCommandExecutor(globals.AppLogger)
		globals.ExecClient = exec.NewClient(mainOSExecutor)
		globals.AssumeYes = assumeYes

		return nil
	},
}

// Execute runs the root command and handles exit codes.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	logLevelAIValue    string
	aiLogFileFlagValue string
	assumeYes          bool
)

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	// Set the version for the --version flag
	rootCmd.Version = globals.AppVersion

	rootCmd.PersistentFlags().
		StringVar(&logLevelAIValue, "log-level-ai", "debug", "AI (JSON) file log level")
	rootCmd.PersistentFlags().
		StringVar(&aiLogFileFlagValue, "ai-log-file", "", "AI (JSON) log file path")
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "Assume 'yes' to all prompts")

	rootCmd.AddCommand(project.ProjectCmd)
	rootCmd.AddCommand(product.ProductCmd)
	rootCmd.AddCommand(factory.FactoryCmd)
	rootCmd.AddCommand(library.LibraryCmd)
	rootCmd.AddCommand(craft.CraftCmd)
	rootCmd.AddCommand(feedback.FeedbackCmd)
	rootCmd.AddCommand(version.VersionCmd)
}

func parseLogLevel(levelStr string, defaultLevel slog.Level) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return defaultLevel
	}
}
