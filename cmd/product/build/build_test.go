// cmd/product/build/build_test.go
package build

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/contextvibes/cli/internal/exec"
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
	ctx context.Context,
	dir string,
	commandName string,
	args ...string,
) (string, string, error) {
	return "", "", errors.New("CaptureOutput not implemented in mock")
}

func (m *mockBuildExecutor) CommandExists(commandName string) bool { return true }

func (m *mockBuildExecutor) Logger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func (m *mockBuildExecutor) UnderlyingExecutor() exec.CommandExecutor { return m }

func setupBuildTest(t *testing.T) (string, *exec.ExecutorClient, *cobra.Command) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { require.NoError(t, os.Chdir(originalWd)) })

	mockExec := &mockBuildExecutor{}
	execClient := exec.NewClient(mockExec)

	// Create a new command instance for each test to avoid state leakage
	cmd := *BuildCmd // Make a copy
	
	// Set up context with dependencies for the command's RunE
	ctx := context.WithValue(context.Background(), "logger", slog.New(slog.DiscardHandler))
	ctx = context.WithValue(ctx, "execClient", execClient)
	cmd.SetContext(ctx)

	return tempDir, execClient, &cmd
}

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
	dummyGoMain := []byte("package main\n\nfunc main() {}\n")

	t.Run("success: standard optimized build", func(t *testing.T) {
		_, execClient, cmd := setupBuildTest(t)
		mockExec := execClient.UnderlyingExecutor().(*mockBuildExecutor)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o644))
		cmdDir := filepath.Join("cmd", "mycoolapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0o644))

		// Reset flags on the command instance for each run
		buildOutputFlag = ""
		buildDebugFlag = false

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
		assert.Equal(t, expectedCommand, mockExec.lastCommand)
	})

	t.Run("success: debug build", func(t *testing.T) {
		_, execClient, cmd := setupBuildTest(t)
		mockExec := execClient.UnderlyingExecutor().(*mockBuildExecutor)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o644))
		cmdDir := filepath.Join("cmd", "myapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0o644))

		buildOutputFlag = ""
		buildDebugFlag = true

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
		assert.Equal(t, expectedCommand, mockExec.lastCommand)
	})
}
