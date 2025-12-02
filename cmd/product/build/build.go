// Package build provides the command to build the project.
package build

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/cmddocs"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed build.md.tpl
var buildLongDescription string

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	buildOutputFlag string
	buildDebugFlag  bool
)

// BuildCmd represents the build command.
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var BuildCmd = &cobra.Command{
	Use: "build [--output <path>] [--debug]",
	Example: `  contextvibes product build                  # Build an optimized binary to ./bin/
  contextvibes product build -o myapp             # Build and name the output 'myapp'
  contextvibes product build --debug              # Build with debug symbols for Delve`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
		ctx := cmd.Context()

		presenter.Summary("Building Go application binary.")

		cwd, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get current working directory: %v", err)

			return fmt.Errorf("failed to get working directory: %w", err)
		}
		projType, err := project.Detect(cwd)
		if err != nil {
			presenter.Error("Failed to detect project type: %v", err)

			return fmt.Errorf("failed to detect project type: %w", err)
		}
		if projType != project.Go {
			presenter.Info("Build command is only applicable for Go projects. Nothing to do.")

			return nil
		}
		presenter.Info("Go project detected.")

		cmdDir := filepath.Join(cwd, "cmd")
		entries, err := os.ReadDir(cmdDir)
		if err != nil {
			if os.IsNotExist(err) {
				presenter.Error(
					"Directory './cmd/' not found. Cannot determine main package to build.",
				)

				//nolint:err113 // Dynamic error is appropriate here.
				return errors.New("cmd directory not found")
			}
			presenter.Error("Failed to read './cmd/' directory: %v", err)

			return fmt.Errorf("failed to read cmd directory: %w", err)
		}

		var mainPackageDirs []string
		for _, entry := range entries {
			if entry.IsDir() {
				mainPackageDirs = append(mainPackageDirs, entry.Name())
			}
		}

		if len(mainPackageDirs) == 0 {
			presenter.Error("No subdirectories found in './cmd/'. Cannot determine main package.")

			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("no main package found in cmd")
		}
		if len(mainPackageDirs) > 1 {
			presenter.Error("Multiple subdirectories found in './cmd/': %v", mainPackageDirs)

			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("ambiguous main package in cmd")
		}
		mainPackageName := mainPackageDirs[0]
		sourcePath := "./" + filepath.ToSlash(filepath.Join("cmd", mainPackageName))
		presenter.Info("Main package found: %s", sourcePath)

		outputPath := buildOutputFlag
		if outputPath == "" {
			binDir := filepath.Join(cwd, "bin")
			//nolint:mnd // 0750 is standard directory permission.
			err := os.MkdirAll(binDir, 0o750)
			if err != nil {
				presenter.Error("Failed to create './bin/' directory: %v", err)

				return fmt.Errorf("failed to create bin directory: %w", err)
			}
			outputPath = filepath.Join("./bin", mainPackageName)
		}
		presenter.Info("Binary will be built to: %s", outputPath)

		buildArgs := []string{"build"}
		if !buildDebugFlag {
			presenter.Info("Compiling optimized binary (without debug symbols).")
			buildArgs = append(buildArgs, "-ldflags", "-s -w")
		} else {
			presenter.Info("Compiling with debug symbols.")
		}
		buildArgs = append(buildArgs, "-o", outputPath, sourcePath)

		presenter.Newline()
		presenter.Step("Running 'go build'...")
		err = globals.ExecClient.Execute(ctx, cwd, "go", buildArgs...)
		if err != nil {
			presenter.Error("'go build' command failed. See output above for details.")

			//nolint:err113 // Dynamic error is appropriate here.
			return errors.New("go build failed")
		}

		presenter.Newline()
		presenter.Success(
			"Build successful. Binary available at: %s",
			presenter.Highlight(outputPath),
		)
		globals.AppLogger.InfoContext(
			ctx,
			"Go build completed successfully",
			"output_path",
			outputPath,
		)

		return nil
	},
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	desc, err := cmddocs.ParseAndExecute(buildLongDescription, nil)
	if err != nil {
		panic(err)
	}

	BuildCmd.Short = desc.Short
	BuildCmd.Long = desc.Long

	BuildCmd.Flags().
		StringVarP(&buildOutputFlag, "output", "o", "", "Output path for the compiled binary.")
	BuildCmd.Flags().
		BoolVar(&buildDebugFlag, "debug", false, "Compile with debug symbols (disables optimization flags).")
}
