package version

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/contextvibes/cli/internal/build"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Info represents the serializable build information.
type Info struct {
	Version   string `json:"version"   yaml:"version"`
	Commit    string `json:"commit"    yaml:"commit"`
	Date      string `json:"date"      yaml:"date"`
	GoVersion string `json:"goVersion" yaml:"goVersion"`
	OS        string `json:"os"        yaml:"os"`
	Arch      string `json:"arch"      yaml:"arch"`
	BuiltBy   string `json:"builtBy"   yaml:"builtBy"`
}

// NewVersionCmd creates and configures the version command.
func NewVersionCmd() *cobra.Command {
	var (
		shortFlag bool
		jsonFlag  bool
		yamlFlag  bool
	)

	//nolint:exhaustruct // Cobra commands rely on zero-value defaults for most fields.
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version and build information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVersion(cmd, shortFlag, jsonFlag, yamlFlag)
		},
	}

	cmd.Flags().BoolVarP(&shortFlag, "short", "s", false, "Print only the version number")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlFlag, "yaml", false, "Output in YAML format")

	return cmd
}

// runVersion handles the execution logic, separated to satisfy funlen linter.
func runVersion(cmd *cobra.Command, short, jsonF, yamlF bool) error {
	// 1. Handle Short Flag (Raw output for scripts)
	if short {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), build.Version)

		return nil
	}

	// 2. Construct Data
	info := Info{
		Version:   build.Version,
		Commit:    build.Commit,
		Date:      build.Date,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		BuiltBy:   build.BuiltBy,
	}

	// 3. Handle JSON Flag
	if jsonF {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")

		if err := enc.Encode(info); err != nil {
			return fmt.Errorf("failed to encode version info to JSON: %w", err)
		}
		return nil
	}

	// 4. Handle YAML Flag
	if yamlF {
		enc := yaml.NewEncoder(cmd.OutOrStdout())

		if err := enc.Encode(info); err != nil {
			return fmt.Errorf("failed to encode version info to YAML: %w", err)
		}
		return nil
	}

	// 5. Default Human-Readable Output (Styled)
	presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())

	presenter.Header("ContextVibes CLI")
	presenter.Detail("Version:    %s", info.Version)
	presenter.Detail("Commit:     %s", info.Commit)
	presenter.Detail("Built:      %s", info.Date)
	presenter.Detail("Go Version: %s", info.GoVersion)
	presenter.Detail("OS/Arch:    %s/%s", info.OS, info.Arch)

	return nil
}
