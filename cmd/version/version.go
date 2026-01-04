// Package version provides the version command.
package version

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// These variables are set via -ldflags during the build process.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "unknown"
)

// Info represents the build information.
type Info struct {
	Version   string `json:"version"   yaml:"version"`
	Commit    string `json:"commit"    yaml:"commit"`
	Date      string `json:"date"      yaml:"date"`
	GoVersion string `json:"goVersion" yaml:"goVersion"`
	OS        string `json:"os"        yaml:"os"`
	Arch      string `json:"arch"      yaml:"arch"`
	BuiltBy   string `json:"builtBy"   yaml:"builtBy"`
}

// VersionCmd represents the version command.
var VersionCmd = NewVersionCmd()

// NewVersionCmd creates and configures the `version` command.
func NewVersionCmd() *cobra.Command {
	var (
		shortFlag bool
		jsonFlag  bool
		yamlFlag  bool
	)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version and build information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// 1. Handle Short Flag
			if shortFlag {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), Version)

				return nil
			}

			// 2. Construct Data
			info := Info{
				Version:   Version,
				Commit:    Commit,
				Date:      Date,
				GoVersion: runtime.Version(),
				OS:        runtime.GOOS,
				Arch:      runtime.GOARCH,
				BuiltBy:   BuiltBy,
			}

			// 3. Handle JSON Flag
			if jsonFlag {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")

				return enc.Encode(info)
			}

			// 4. Handle YAML Flag
			if yamlFlag {
				enc := yaml.NewEncoder(cmd.OutOrStdout())

				return enc.Encode(info)
			}

			// 5. Default Human-Readable Output
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ContextVibes CLI\n")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Version:    %s\n", info.Version)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Commit:     %s\n", info.Commit)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Built:      %s\n", info.Date)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Go Version: %s\n", info.GoVersion)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  OS/Arch:    %s/%s\n", info.OS, info.Arch)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&shortFlag, "short", "s", false, "Print only the version number")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlFlag, "yaml", false, "Output in YAML format")

	return cmd
}
