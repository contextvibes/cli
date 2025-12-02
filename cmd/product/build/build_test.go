// Package build_test contains tests for the build command.
package build_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/contextvibes/cli/cmd/product/build"
	"github.com/contextvibes/cli/internal/exec"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBuildExecutor struct {
	ExecuteFunc func(ctx context.Context, dir string, commandName string, args ...string) error
	lastCommand []string
}

func (m *mockBuildExecutor) Execute(
	ctx context.Context,
	dir string,
	commandName string,
	args ...string,
) error {
	m.lastCommand = append([]string{commandName}, args...)
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, dir, commandName, args...)
	}

	return nil
}

func (m *mockBuildExecutor) CaptureOutput(
	_ context.Context,
	_ string,
	_ string,
	_ ...string,
) (string, string, error) {
	//nolint:err113 // Dynamic error is appropriate here.
	return "", "", errors.New("CaptureOutput not implemented in mock")
}

//nolint:revive // Unused parameter is expected in mock.
func (m *mockBuildExecutor) CommandExists(commandName string) bool { return true }

func (m *mockBuildExecutor) Logger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

//nolint:ireturn // Returning interface is required for mock.
func (m *mockBuildExecutor) UnderlyingExecutor() exec.CommandExecutor { return m }

//nolint:unparam // Return values are used in tests.
func setupBuildTest(t *testing.T) (string, *exec.ExecutorClient, *cobra.Command) {
	t.Helper()
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	//nolint:usetesting // os.Chdir is required for test setup.
	require.NoError(t, os.Chdir(tempDir))
	//nolint:usetesting // os.Chdir is required for test setup.
	t.Cleanup(func() { require.NoError(t, os.Chdir(originalWd)) })

	//nolint:exhaustruct // Mock executor partial initialization is fine.
	mockExec := &mockBuildExecutor{}
	execClient := exec.NewClient(mockExec)

	globals.ExecClient = execClient
	globals.AppLogger = slog.New(slog.DiscardHandler)

	// Create a new command instance for each test to avoid state leakage
	cmd := *build.BuildCmd // Make a copy
	cmd.SetContext(context.Background())

	return tempDir, execClient, &cmd
}

//nolint:unparam // Return values are used in tests.
func runBuildCmd(cmd *cobra.Command, args []string) (string, string, error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)

	// Reset flags manually because we are using a global flag variable in the command implementation
	// This is a workaround for testing Cobra commands with global flags.
	// In a real refactor, we should move flags to a struct.
	// For now, we rely on the fact that we are running sequentially (no t.Parallel).
	_ = cmd.Flags().Set("output", "")
	_ = cmd.Flags().Set("debug", "false")

	err := cmd.Execute() // Use Execute instead of RunE to ensure full Cobra lifecycle including flag parsing

	return outBuf.String(), errBuf.String(), err
}

//nolint:paralleltest // BuildCmd uses global flags which are not thread-safe.
func TestBuildCmd(t *testing.T) {
	dummyGoMain := []byte("package main\n\nfunc main() {}\n")

	//nolint:paralleltest // BuildCmd uses global flags which are not thread-safe.
	t.Run("success: standard optimized build", func(t *testing.T) {
		_, execClient, cmd := setupBuildTest(t)
		underlyingMock, ok := execClient.UnderlyingExecutor().(*mockBuildExecutor)
		require.True(t, ok)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o600))

		cmdDir := filepath.Join("cmd", "mycoolapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0o600))

		out, _, err := runBuildCmd(cmd, []string{})
		require.NoError(t, err)

		assert.Contains(t, out, "Build successful")

		expectedCommand := []string{
			"go",
			"build",
			"-ldflags",
			"-s -w",
			"-o",
			filepath.Join("bin", "mycoolapp"),
			"./" + filepath.ToSlash(filepath.Join("cmd", "mycoolapp")),
		}
		assert.Equal(t, expectedCommand, underlyingMock.lastCommand)
	})

	//nolint:paralleltest // BuildCmd uses global flags which are not thread-safe.
	t.Run("success: debug build", func(t *testing.T) {
		_, execClient, cmd := setupBuildTest(t)
		underlyingMock, ok := execClient.UnderlyingExecutor().(*mockBuildExecutor)
		require.True(t, ok)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o600))

		cmdDir := filepath.Join("cmd", "myapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0o600))

		out, _, err := runBuildCmd(cmd, []string{"--debug"}) // Pass flag to args
		require.NoError(t, err)

		assert.Contains(t, out, "Compiling with debug symbols.")

		expectedCommand := []string{
			"go",
			"build",
			"-o",
			filepath.Join("bin", "myapp"),
			"./" + filepath.ToSlash(filepath.Join("cmd", "myapp")),
		}
		assert.Equal(t, expectedCommand, underlyingMock.lastCommand)
	})
}
