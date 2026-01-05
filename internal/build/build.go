// Package build holds the build-time metadata injected by the linker.
package build

// These variables are set via -ldflags during the build process.
// If not set via flags (e.g. during 'go run'), they default to "dev" or "unknown".
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)
