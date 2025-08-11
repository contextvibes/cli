// FILE: cmd/build.go
package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/contextvibes/cli/internal/project"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	buildOutputFlag string
	buildDebugFlag  bool
)

var buildCmd = &cobra.Command{
	Use:   "build [--output <path>] [--debug]",
	Short: "Compiles the Go project's main application.",
	Long: `Detects a Go project and compiles its main application.

By default, this command looks for a single subdirectory within the 'cmd/' directory
to determine the main package to build. It produces an optimized, stripped binary
in the './bin/' directory.

Use the --debug flag to compile with debugging symbols included.`,
	Example: `  contextvibes build                  # Build an optimized binary to ./bin/
  contextvibes build -o myapp             # Build and name the output 'myapp'
  contextvibes build --debug              # Build with debug symbols for Delve`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// CORRECTED: Use the command's configured output streams. This makes the command testable.
		presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr(), os.Stdin)
		logger := AppLogger
		ctx := cmd.Context()

		presenter.Summary("Building Go application binary.")

		cwd, err := os.Getwd()
		if err != nil {
			presenter.Error("Failed to get current working directory: %v", err)
			return err
		}
		projType, err := project.Detect(cwd)
		if err != nil {
			presenter.Error("Failed to detect project type: %v", err)
			return err
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
				presenter.Error("Directory './cmd/' not found. Cannot determine main package to build.")
				presenter.Advice("This command expects a conventional Go project layout with a './cmd/<appname>/' directory.")
				return errors.New("cmd directory not found")
			}
			presenter.Error("Failed to read './cmd/' directory: %v", err)
			return err
		}

		var mainPackageDirs []string
		for _, entry := range entries {
			if entry.IsDir() {
				mainPackageDirs = append(mainPackageDirs, entry.Name())
			}
		}

		if len(mainPackageDirs) == 0 {
			presenter.Error("No subdirectories found in './cmd/'. Cannot determine main package.")
			return errors.New("no main package found in cmd")
		}
		if len(mainPackageDirs) > 1 {
			presenter.Error("Multiple subdirectories found in './cmd/': %v", mainPackageDirs)
			presenter.Advice("Unsure which application to build. Please ensure only one main package exists in './cmd/'.")
			return errors.New("ambiguous main package in cmd")
		}
		mainPackageName := mainPackageDirs[0]
		sourcePath := "./" + filepath.ToSlash(filepath.Join("cmd", mainPackageName))
		presenter.Info("Main package found: %s", sourcePath)

		outputPath := buildOutputFlag
		if outputPath == "" {
			binDir := filepath.Join(cwd, "bin")
			if err := os.MkdirAll(binDir, 0750); err != nil {
				presenter.Error("Failed to create './bin/' directory: %v", err)
				return err
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
		err = ExecClient.Execute(ctx, cwd, "go", buildArgs...)
		if err != nil {
			presenter.Error("'go build' command failed. See output above for details.")
			return errors.New("go build failed")
		}

		presenter.Newline()
		presenter.Success("Build successful. Binary available at: %s", presenter.Highlight(outputPath))
		logger.InfoContext(ctx, "Go build completed successfully", "output_path", outputPath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&buildOutputFlag, "output", "o", "", "Output path for the compiled binary.")
	buildCmd.Flags().BoolVar(&buildDebugFlag, "debug", false, "Compile with debug symbols (disables optimization flags).")
}
