// internal/globals/globals.go
package globals

import (
	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/exec"
	"log/slog"
)

// These variables are initialized by the rootCmd in cmd/root.go
var (
	AppLogger       *slog.Logger
	LoadedAppConfig *config.Config
	ExecClient      *exec.ExecutorClient
	AssumeYes       bool
	AppVersion      string
)
