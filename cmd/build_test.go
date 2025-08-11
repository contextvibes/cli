// cmd/build_test.go
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExecutor for build command tests.
type mockBuildExecutor struct {
	ExecuteFunc func(ctx context.Context, dir string, commandName string, args ...string) error
	// Store the last command that was executed for inspection
	lastCommand []string
}

func (m *mockBuildExecutor) Execute(ctx context.Context, dir string, commandName string, args ...string) error {
	m.lastCommand = append([]string{commandName}, args...)
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, dir, commandName, args...)
	}

	return nil // Default to success
}
func (m *mockBuildExecutor) CaptureOutput(ctx context.Context, dir string, commandName string, args ...string) (string, string, error) {
	return "", "", errors.New("CaptureOutput not implemented in mock")
}
func (m *mockBuildExecutor) CommandExists(commandName string) bool { return true }
func (m *mockBuildExecutor) Logger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
func (m *mockBuildExecutor) UnderlyingExecutor() exec.CommandExecutor { return m }

// setupBuildTest is a helper to initialize a test environment.
func setupBuildTest(t *testing.T) (string, *bytes.Buffer, *mockBuildExecutor) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	// Change into the temp directory for the duration of the test
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { require.NoError(t, os.Chdir(originalWd)) })

	// Setup mock executor and global variables
	mockExec := &mockBuildExecutor{}
	originalExecClient := ExecClient
	ExecClient = exec.NewClient(mockExec)

	t.Cleanup(func() { ExecClient = originalExecClient })

	originalLogger := AppLogger
	AppLogger = slog.New(slog.DiscardHandler)

	t.Cleanup(func() { AppLogger = originalLogger })

	outputBuffer := new(bytes.Buffer)

	return tempDir, outputBuffer, mockExec
}

func TestBuildCmd(t *testing.T) {
	t.Run("success: standard optimized build", func(t *testing.T) {
		_, out, mockExec := setupBuildTest(t)

		// Create a valid Go project structure
		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join("cmd", "mycoolapp"), 0755))

		rootCmd.SetArgs([]string{"build"})
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		err := rootCmd.Execute()
		require.NoError(t, err)

		// Assertions
		output := out.String()
		assert.Contains(t, output, "Building Go application binary.")
		assert.Contains(t, output, "Go project detected.")
		assert.Contains(t, output, "Main package found: cmd/mycoolapp")
		assert.Contains(t, output, "Compiling optimized binary")
		assert.Contains(t, output, "Build successful")
		assert.Contains(t, output, filepath.Join("bin", "mycoolapp"))

		expectedCommand := []string{"go", "build", "-ldflags", "-s -w", "-o", filepath.Join("bin", "mycoolapp"), filepath.Join("cmd", "mycoolapp")}
		assert.Equal(t, expectedCommand, mockExec.lastCommand)
	})

	t.Run("success: debug build", func(t *testing.T) {
		_, out, mockExec := setupBuildTest(t)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join("cmd", "myapp"), 0755))

		rootCmd.SetArgs([]string{"build", "--debug"})
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		err := rootCmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Compiling with debug symbols.")

		expectedCommand := []string{"go", "build", "-o", filepath.Join("bin", "myapp"), filepath.Join("cmd", "myapp")}
		assert.Equal(t, expectedCommand, mockExec.lastCommand)
		// Ensure optimization flags are NOT present
		assert.NotContains(t, mockExec.lastCommand, "-ldflags")
	})

	t.Run("success: custom output", func(t *testing.T) {
		_, out, mockExec := setupBuildTest(t)

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join("cmd", "mytool"), 0755))

		rootCmd.SetArgs([]string{"build", "-o", "dist/mytool.exe"})
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		err := rootCmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Binary will be built to: dist/mytool.exe")

		expectedCommand := []string{"go", "build", "-ldflags", "-s -w", "-o", "dist/mytool.exe", filepath.Join("cmd", "mytool")}
		assert.Equal(t, expectedCommand, mockExec.lastCommand)
	})

	t.Run("failure: not a go project", func(t *testing.T) {
		_, out, mockExec := setupBuildTest(t)
		// DO NOT create go.mod

		rootCmd.SetArgs([]string{"build"})
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		err := rootCmd.Execute()
		require.NoError(t, err) // Command should exit gracefully

		assert.Contains(t, out.String(), "Build command is only applicable for Go projects.")
		assert.Nil(t, mockExec.lastCommand, "No build command should have been executed")
	})

	t.Run("failure: no cmd directory", func(t *testing.T) {
		_, out, mockExec := setupBuildTest(t)
		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))

		rootCmd.SetArgs([]string{"build"})
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		err := rootCmd.Execute()
		require.Error(t, err)

		assert.Contains(t, out.String(), "Directory './cmd/' not found.")
		assert.Nil(t, mockExec.lastCommand, "No build command should have been executed")
	})

	t.Run("failure: ambiguous cmd directory", func(t *testing.T) {
		_, out, mockExec := setupBuildTest(t)
		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join("cmd", "app1"), 0755))
		require.NoError(t, os.MkdirAll(filepath.Join("cmd", "app2"), 0755))

		rootCmd.SetArgs([]string{"build"})
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		err := rootCmd.Execute()
		require.Error(t, err)

		assert.Contains(t, out.String(), "Multiple subdirectories found in './cmd/'")
		assert.Nil(t, mockExec.lastCommand, "No build command should have been executed")
	})

	t.Run("failure: go build command fails", func(t *testing.T) {
		_, out, mockExec := setupBuildTest(t)

		mockExec.ExecuteFunc = func(ctx context.Context, dir, cmd string, args ...string) error {
			return errors.New("simulated compilation error")
		}

		require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join("cmd", "failingapp"), 0755))

		rootCmd.SetArgs([]string{"build"})
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)

		err := rootCmd.Execute()
		require.Error(t, err)
		assert.Contains(t, out.String(), "'go build' command failed.")
	})
}
