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

var errCaptureOutputNotImplemented = errors.New("CaptureOutput not implemented in mock")

type mockBuildExecutor struct {
	ExecuteFunc func(ctx context.Context, dir, commandName string, args ...string) error
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
	return "", "", errCaptureOutputNotImplemented
}

func (m *mockBuildExecutor) CommandExists(_ string) bool { return true }

func (m *mockBuildExecutor) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func (m *mockBuildExecutor) UnderlyingExecutor() *mockBuildExecutor { return m }

func setupBuildTest(t *testing.T) (*exec.ExecutorClient, *cobra.Command) {
	t.Helper()

	t.Chdir(t.TempDir())

	mockExec := &mockBuildExecutor{
		ExecuteFunc: nil,
		lastCommand: nil,
	}
	execClient := exec.NewClient(mockExec)

	globals.ExecClient = execClient
	globals.AppLogger = slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Create a new command instance for each test to avoid state leakage
	cmd := *build.BuildCmd // Make a copy
	cmd.SetContext(context.Background())

	return execClient, &cmd
}

func runBuildCmd(cmd *cobra.Command, args []string) (string, error) {
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

	return outBuf.String(), err
}

func TestBuildCmd(t *testing.T) {
	t.Parallel()

	dummyGoMain := []byte("package main\n\nfunc main() {}\n")

	t.Run("success: standard optimized build", func(t *testing.T) {
		t.Parallel()

		execClient, cmd := setupBuildTest(t)
		underlyingMock, ok := execClient.UnderlyingExecutor().(*mockBuildExecutor)
		require.True(t, ok)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o600))

		cmdDir := filepath.Join("cmd", "mycoolapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0o600))

		out, err := runBuildCmd(cmd, []string{})
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

		execClient, cmd := setupBuildTest(t)
		underlyingMock, ok := execClient.UnderlyingExecutor().(*mockBuildExecutor)
		require.True(t, ok)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o600))

		cmdDir := filepath.Join("cmd", "myapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0o600))

		out, err := runBuildCmd(cmd, []string{"--debug"}) // Pass flag to args
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
