package version_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/contextvibes/cli/cmd/version"
	"github.com/contextvibes/cli/internal/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmd(t *testing.T) {
	// Save original state
	origVersion := build.Version
	origCommit := build.Commit
	origDate := build.Date

	// Restore state after test
	t.Cleanup(func() {
		build.Version = origVersion
		build.Commit = origCommit
		build.Date = origDate
	})

	// Set test state
	build.Version = "1.2.3-test"
	build.Commit = "abcdef"
	build.Date = "2025-01-01"

	t.Run("default output", func(t *testing.T) {
		cmd := version.NewVersionCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "ContextVibes CLI")
		assert.Contains(t, output, "Version:    1.2.3-test")
		assert.Contains(t, output, "Commit:     abcdef")
	})

	t.Run("short output", func(t *testing.T) {
		cmd := version.NewVersionCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--short"})

		err := cmd.Execute()
		require.NoError(t, err)

		output := strings.TrimSpace(buf.String())
		assert.Equal(t, "1.2.3-test", output)
	})

	t.Run("json output", func(t *testing.T) {
		cmd := version.NewVersionCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--json"})

		err := cmd.Execute()
		require.NoError(t, err)

		var info version.Info
		err = json.Unmarshal(buf.Bytes(), &info)
		require.NoError(t, err)

		assert.Equal(t, "1.2.3-test", info.Version)
		assert.Equal(t, "abcdef", info.Commit)
	})
}
