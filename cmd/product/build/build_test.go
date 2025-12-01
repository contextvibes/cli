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
	return "", "", errors.New("CaptureOutput not implemented in mock")
}

//nolint:revive // Unused parameter is expected in mock.
func (m *mockBuildExecutor) CommandExists(commandName string) bool { return true }

func (m *mockBuildExecutor) Logger() *slog.Logger {
	// THE FIX: Return a valid logger that discards output.
	return slog.New(slog.DiscardHandler)
}

//nolint:ireturn // Returning interface is required for mock.
func (m *mockBuildExecutor) UnderlyingExecutor() exec.CommandExecutor { return m }

func setupBuildTest(t *testing.T) (string, *exec.ExecutorClient, *cobra.Command) {
	t.Helper()
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { require.NoError(t, os.Chdir(originalWd)) })

	//nolint:exhaustruct // Mock executor partial initialization is fine.
	mockExec := &mockBuildExecutor{}
	execClient := exec.NewClient(mockExec)

	globals.ExecClient = execClient
	// THE FIX: Initialize the global logger with a discard handler for tests.
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

	err := cmd.RunE(cmd, args)

	return outBuf.String(), errBuf.String(), err
}

func TestBuildCmd(t *testing.T) {
	t.Parallel()

	dummyGoMain := []byte("package main\n\nfunc main() {}\n")

	t.Run("success: standard optimized build", func(t *testing.T) {
		t.Parallel()
		_, execClient, cmd := setupBuildTest(t)
		underlyingMock, ok := execClient.UnderlyingExecutor().(*mockBuildExecutor)
		require.True(t, ok)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o600))

		cmdDir := filepath.Join("cmd", "mycoolapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0o600))

		// Reset flags on the command instance for each run
		// Note: Since we are copying the command struct, we might need to reset flags if they are global
		// But in this test setup we are relying on the fact that we are running in parallel and
		// ideally flags should not be global. However, BuildCmd uses global flags.
		// This is a limitation of the current CLI structure.
		// For now, we assume sequential execution or isolated process for robust testing of globals.
		// But since we added t.Parallel(), we might have race conditions on globals.
		// TODO: Refactor CLI to avoid global flags for better testability.

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

	t.Run("success: debug build", func(t *testing.T) {
		t.Parallel()
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
