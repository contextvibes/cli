// FILE: internal/kickoff/orchestrator_test.go
package kickoff

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock Presenter that satisfies PresenterInterface ---
type MockPresenter struct {
	mock.Mock
	OutputBuffer *bytes.Buffer
}

func NewMockPresenter() *MockPresenter {
	return &MockPresenter{OutputBuffer: new(bytes.Buffer)}
}

func (m *MockPresenter) Header(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "HEADER: "+format+"\n", a...)
}
func (m *MockPresenter) Summary(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "SUMMARY: "+format+"\n", a...)
}
func (m *MockPresenter) Step(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "STEP: "+format+"\n", a...)
}
func (m *MockPresenter) Info(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "INFO: "+format+"\n", a...)
}
func (m *MockPresenter) Success(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "SUCCESS: "+format+"\n", a...)
}
func (m *MockPresenter) Error(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "ERROR: "+format+"\n", a...)
}
func (m *MockPresenter) Warning(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "WARNING: "+format+"\n", a...)
}
func (m *MockPresenter) Advice(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "ADVICE: "+format+"\n", a...)
}
func (m *MockPresenter) Detail(format string, a ...any) {
	fmt.Fprintf(m.OutputBuffer, "DETAIL: "+format+"\n", a...)
}
func (m *MockPresenter) Highlight(text string) string { return "*" + text + "*" }
func (m *MockPresenter) Newline()                     { fmt.Fprintln(m.OutputBuffer) }
func (m *MockPresenter) PromptForInput(prompt string) (string, error) {
	args := m.Called(prompt)
	return args.String(0), args.Error(1)
}
func (m *MockPresenter) PromptForConfirmation(prompt string) (bool, error) {
	args := m.Called(prompt)
	return args.Bool(0), args.Error(1)
}

func setupTestOrchestrator(t *testing.T) (*Orchestrator, *config.Config, *MockPresenter, string) {
	logger := slog.New(slog.DiscardHandler)
	appCfg := config.GetDefaultConfig()
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, ".contextvibes.yaml")
	require.NoError(t, config.UpdateAndSaveConfig(appCfg, configFilePath))

	mockPresenter := NewMockPresenter()
	var gitClientForOrc *git.GitClient = nil

	orc := NewOrchestrator(logger, appCfg, mockPresenter, gitClientForOrc, configFilePath, false)
	require.NotNil(t, orc)

	return orc, appCfg, mockPresenter, configFilePath
}

func TestExecuteKickoff_ModeSelection(t *testing.T) {
	ctx := context.Background()

	t.Run("strategic flag forces strategic mode", func(t *testing.T) {
		orc, _, mockPresenter, _ := setupTestOrchestrator(t)
		// Mock the interactive session
		mockPresenter.On("Info", mock.Anything).Return()
		mockPresenter.On("Summary", mock.Anything).Return()
		mockPresenter.On("Step", mock.Anything).Return()
		mockPresenter.On("Success", mock.Anything).Return()
		mockPresenter.On("Header", mock.Anything).Return()
		mockPresenter.On("Newline").Return()
		mockPresenter.On("PromptForInput", mock.Anything).Return("", nil)
		mockPresenter.On("PromptForConfirmation", mock.Anything).Return(true, nil)

		err := orc.ExecuteKickoff(ctx, true, "")
		require.NoError(t, err)
		assert.Contains(t, mockPresenter.OutputBuffer.String(), "ContextVibes: Strategic Project Kickoff - Prompt Generation")
	})
}
