// FILE: cmd/build_test.go
package cmd

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

// mockBuildExecutor remains the same
type mockBuildExecutor struct {
	ExecuteFunc func(ctx context.Context, dir string, commandName string, args ...string) error
	lastCommand []string
}

func (m *mockBuildExecutor) Execute(ctx context.Context, dir string, commandName string, args ...string) error {
	m.lastCommand = append([]string{commandName}, args...)
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, dir, commandName, args...)
	}
	return nil
}
func (m *mockBuildExecutor) CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (string, string, error) {
	return "", "", errors.New("CaptureOutput not implemented in mock")
}
func (m *mockBuildExecutor) CommandExists(commandName string) bool { return true }
func (m *mockBuildExecutor) Logger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
func (m *mockBuildExecutor) UnderlyingExecutor() exec.CommandExecutor { return m }

// setupBuildTest remains mostly the same
func setupBuildTest(t *testing.T) (string, *mockBuildExecutor) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { require.NoError(t, os.Chdir(originalWd)) })

	mockExec := &mockBuildExecutor{}
	// Set the global ExecClient which will be used by the command's RunE function
	ExecClient = exec.NewClient(mockExec)
	t.Cleanup(func() { ExecClient = nil }) // Clean up global state

	AppLogger = slog.New(slog.DiscardHandler)
	t.Cleanup(func() { AppLogger = nil })

	return tempDir, mockExec
}

// Helper to run the build command's logic for tests
func runBuildCmd(cmd *cobra.Command, args []string) (string, string, error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)

	// Directly call the command's execution logic
	err := buildCmd.RunE(cmd, args)

	return outBuf.String(), errBuf.String(), err
}

func TestBuildCmd(t *testing.T) {
	dummyGoMain := []byte("package main\n\nfunc main() {}\n")

	t.Run("success: standard optimized build", func(t *testing.T) {
		_, mockExec := setupBuildTest(t)
		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		cmdDir := filepath.Join("cmd", "mycoolapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0644))

		// Reset flags for each run
		buildOutputFlag = ""
		buildDebugFlag = false

		out, _, err := runBuildCmd(buildCmd, []string{})
		require.NoError(t, err)

		assert.Contains(t, out, "Build successful")
		expectedCommand := []string{"go", "build", "-ldflags", "-s -w", "-o", filepath.Join("bin", "mycoolapp"), "./" + filepath.ToSlash(filepath.Join("cmd", "mycoolapp"))}
		assert.Equal(t, expectedCommand, mockExec.lastCommand)
	})

	t.Run("success: debug build", func(t *testing.T) {
		_, mockExec := setupBuildTest(t)
		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		cmdDir := filepath.Join("cmd", "myapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0644))

		buildOutputFlag = ""
		buildDebugFlag = true // Set the flag for debug

		out, _, err := runBuildCmd(buildCmd, []string{})
		require.NoError(t, err)

		assert.Contains(t, out, "Compiling with debug symbols.")
		expectedCommand := []string{"go", "build", "-o", filepath.Join("bin", "myapp"), "./" + filepath.ToSlash(filepath.Join("cmd", "myapp"))}
		assert.Equal(t, expectedCommand, mockExec.lastCommand)
	})

	t.Run("failure: not a go project", func(t *testing.T) {
		_, _ = setupBuildTest(t)
		buildOutputFlag = ""
		buildDebugFlag = false

		out, _, err := runBuildCmd(buildCmd, []string{})
		require.NoError(t, err)
		assert.Contains(t, out, "Build command is only applicable for Go projects.")
	})

	t.Run("failure: go build command fails", func(t *testing.T) {
		_, mockExec := setupBuildTest(t)
		mockExec.ExecuteFunc = func(ctx context.Context, dir, cmd string, args ...string) error {
			return errors.New("simulated compilation error")
		}
		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		cmdDir := filepath.Join("cmd", "failingapp")
		require.NoError(t, os.MkdirAll(cmdDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), dummyGoMain, 0644))

		buildOutputFlag = ""
		buildDebugFlag = false

		_, _, err := runBuildCmd(buildCmd, []string{})
		require.Error(t, err)
		assert.Equal(t, "go build failed", err.Error())
	})
}
