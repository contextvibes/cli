// Package globals provides global variables for the application.
package globals

import (
	"log/slog"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
)

// These variables are initialized by the rootCmd in cmd/root.go
//
//nolint:gochecknoglobals // Global state is required for CLI initialization.
var (
	AppLogger       *slog.Logger
	LoadedAppConfig *config.Config
	ExecClient      *exec.ExecutorClient
	AssumeYes       bool
	// AppVersion is the current version of the CLI.
	AppVersion = "0.4.1"
)
