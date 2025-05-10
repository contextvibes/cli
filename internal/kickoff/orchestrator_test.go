package kickoff

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	// "strings" // Only re-add if specific string manipulations are needed in tests

	"github.com/contextvibes/cli/internal/config"
	"github.com/contextvibes/cli/internal/git" 
	"github.com/contextvibes/cli/internal/ui"  
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- SourcedInputReader Helper ---
type SourcedInputReader struct {
	inputs []string
	index  int
}

func NewSourcedInputReader(inputs []string) *SourcedInputReader {
	return &SourcedInputReader{inputs: inputs}
}
func (sir *SourcedInputReader) Read(p []byte) (n int, err error) {
	if sir.index >= len(sir.inputs) {
		return 0, io.EOF 
	}
	inputLine := sir.inputs[sir.index] + "\n" 
	sir.index++
	return copy(p, []byte(inputLine)), nil
}

// --- Mock Presenter ---
type MockPresenter struct {
	mock.Mock 
	OutputBuffer   *bytes.Buffer   
	ErrorBuffer    *bytes.Buffer   
	Confirmations  []bool          
	confirmIndex   int
	InputResponses []string        
	inputIndex     int
}

func NewMockPresenter() *MockPresenter {
	return &MockPresenter{
		OutputBuffer: new(bytes.Buffer),
		ErrorBuffer:  new(bytes.Buffer),
	}
}
func (m *MockPresenter) Out() io.Writer { return m.OutputBuffer }
func (m *MockPresenter) Err() io.Writer { return m.ErrorBuffer }
func (m *MockPresenter) SetInput(inputs []string) { m.InputResponses = inputs; m.inputIndex = 0 }
func (m *MockPresenter) SetConfirmations(confirmations []bool) { m.Confirmations = confirmations; m.confirmIndex = 0 }
func (m *MockPresenter) PromptForInput(prompt string) (string, error) {
	m.Called(prompt) 
	m.ErrorBuffer.WriteString(prompt) 
	if m.inputIndex < len(m.InputResponses) {
		input := m.InputResponses[m.inputIndex]; m.inputIndex++; return input, nil
	}
	return "", errors.New("mock presenter: no more inputs available")
}
func (m *MockPresenter) PromptForConfirmation(prompt string) (bool, error) {
	m.Called(prompt) 
	m.ErrorBuffer.WriteString(prompt) 
	if m.confirmIndex < len(m.Confirmations) {
		confirmation := m.Confirmations[m.confirmIndex]; m.confirmIndex++; return confirmation, nil
	}
	return false, errors.New("mock presenter: no more confirmations available")
}
func (m *MockPresenter) Header(format string, a ...any)  { fmt.Fprintf(m.OutputBuffer, "HEADER: "+format+"\n", a...) }
func (m *MockPresenter) Summary(format string, a ...any) { fmt.Fprintf(m.OutputBuffer, "SUMMARY: "+format+"\n", a...) }
func (m *MockPresenter) Step(format string, a ...any)    { fmt.Fprintf(m.OutputBuffer, "STEP: "+format+"\n", a...) }
func (m *MockPresenter) Info(format string, a ...any)    { fmt.Fprintf(m.OutputBuffer, "INFO: "+format+"\n", a...) }
func (m *MockPresenter) Success(format string, a ...any) { fmt.Fprintf(m.OutputBuffer, "SUCCESS: "+format+"\n", a...) }
func (m *MockPresenter) Error(format string, a ...any)   { fmt.Fprintf(m.ErrorBuffer, "ERROR: "+format+"\n", a...) }
func (m *MockPresenter) Warning(format string, a ...any) { fmt.Fprintf(m.ErrorBuffer, "WARNING: "+format+"\n", a...) }
func (m *MockPresenter) Advice(format string, a ...any)  { fmt.Fprintf(m.OutputBuffer, "ADVICE: "+format+"\n", a...) }
func (m *MockPresenter) Detail(format string, a ...any)  { fmt.Fprintf(m.OutputBuffer, "DETAIL: "+format+"\n", a...) }
func (m *MockPresenter) Highlight(text string) string    { return "*" + text + "*" }
func (m *MockPresenter) Newline()                        { fmt.Fprintln(m.OutputBuffer) }
func (m *MockPresenter) InfoPrefixOnly()                 { fmt.Fprint(m.OutputBuffer, "INFO: ") }
func (m *MockPresenter) Separator()                      { fmt.Fprintln(m.OutputBuffer, "---SEPARATOR---") }

// --- Mock GitClient ---
type MockGitClient struct { mock.Mock }
func (m *MockGitClient) GetCurrentBranchName(ctx context.Context) (string, error) { args := m.Called(ctx); return args.String(0), args.Error(1) }
func (m *MockGitClient) IsWorkingDirClean(ctx context.Context) (bool, error)      { args := m.Called(ctx); return args.Bool(0), args.Error(1) }
func (m *MockGitClient) PullRebase(ctx context.Context, branch string) error       { args := m.Called(ctx, branch); return args.Error(0) }
func (m *MockGitClient) LocalBranchExists(ctx context.Context, branchName string) (bool, error) { args := m.Called(ctx, branchName); return args.Bool(0), args.Error(1) }
func (m *MockGitClient) CreateAndSwitchBranch(ctx context.Context, newBranch, baseBranch string) error { args := m.Called(ctx, newBranch, baseBranch); return args.Error(0) }
func (m *MockGitClient) PushAndSetUpstream(ctx context.Context, branchName string) error { args := m.Called(ctx, branchName); return args.Error(0) }

func setupTestOrchestrator(t *testing.T) (
	orc *Orchestrator, 
	appCfg *config.Config, 
	presenterOut *bytes.Buffer, 
	presenterErr *bytes.Buffer, 
	inputReader *SourcedInputReader, 
	mockGit *MockGitClient,      
	configFilePath string,
) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) 
	appCfg = config.GetDefaultConfig()                       
	
	tempDir := t.TempDir()
	configFilePath = filepath.Join(tempDir, ".contextvibes.yaml")

	err := config.UpdateAndSaveConfig(appCfg, configFilePath)
	require.NoError(t, err, "Failed to write initial dummy config for test setup")

	presenterOut = new(bytes.Buffer)
	presenterErr = new(bytes.Buffer)
	inputReader = NewSourcedInputReader(nil) 

	presenter := ui.NewPresenter(presenterOut, presenterErr, inputReader)
	mockGit = new(MockGitClient)
	
	var gitClientForOrc *git.GitClient = nil 
	
	// Pass a default 'false' for assumeYes to NewOrchestrator
	orc = NewOrchestrator(logger, appCfg, presenter, gitClientForOrc, configFilePath, false /* globalAssumeYes */)
	require.NotNil(t, orc)
	
	return orc, appCfg, presenterOut, presenterErr, inputReader, mockGit, configFilePath
}


func TestNewOrchestrator(t *testing.T) {
	orc, cfg, _, _, _, _, configPath := setupTestOrchestrator(t) 
	assert.NotNil(t, orc.logger)
	assert.Equal(t, cfg, orc.config)
	assert.NotNil(t, orc.presenter) 
	assert.Equal(t, configPath, orc.configFilePath) 
	assert.False(t, orc.assumeYes, "assumeYes should be false by default from setupTestOrchestrator")
}

func TestExecuteKickoff_ModeSelection(t *testing.T) {
	ctx := context.Background()

	strategicSetupInputs := []string{
		"bash_cat_eof", "raw_markdown", 
		"TestProjectStrategic", "Go CLI For Strategic", "new", 
		"y", "y", 
	}

	t.Run("strategic flag forces strategic mode", func(t *testing.T) {
		orc, _, outBuf, _, inputReader, _, _ := setupTestOrchestrator(t)
		inputReader.inputs = strategicSetupInputs
		orc.assumeYes = false // Explicitly set for this test case if needed, or rely on setupTestOrchestrator default
		
		// ExecuteKickoff no longer takes globalAssumeYes
		err := orc.ExecuteKickoff(ctx, true /*isStrategicFlag*/, "") 
		require.NoError(t, err) 
		assert.Contains(t, outBuf.String(), "ContextVibes: Strategic Project Kickoff - Prompt Generation")
	})

	t.Run("no strategic flag, kickoff not completed, runs strategic", func(t *testing.T) {
		orc, cfg, outBuf, _, inputReader, _, _ := setupTestOrchestrator(t)
		defaultFalse := false
		cfg.ProjectState.StrategicKickoffCompleted = &defaultFalse
		orc.assumeYes = false

		inputReader.inputs = strategicSetupInputs
		
		err := orc.ExecuteKickoff(ctx, false /*isStrategicFlag*/, "")
		require.NoError(t, err)
		assert.Contains(t, outBuf.String(), "No prior strategic kickoff detected")
		assert.Contains(t, outBuf.String(), "ContextVibes: Strategic Project Kickoff - Prompt Generation")
	})

	t.Run("no strategic flag, kickoff nil (default), runs strategic", func(t *testing.T) {
		orc, cfg, outBuf, _, inputReader, _, _ := setupTestOrchestrator(t)
		cfg.ProjectState.StrategicKickoffCompleted = nil 
		orc.assumeYes = false

		inputReader.inputs = strategicSetupInputs
		
		err := orc.ExecuteKickoff(ctx, false /*isStrategicFlag*/, "")
		require.NoError(t, err)
		assert.Contains(t, outBuf.String(), "No prior strategic kickoff detected") 
		assert.Contains(t, outBuf.String(), "ContextVibes: Strategic Project Kickoff - Prompt Generation")
	})

	t.Run("no strategic flag, kickoff completed, tries daily, fails due to nil gitClient", func(t *testing.T) {
		orc, cfg, outBuf, _, _, _, _ := setupTestOrchestrator(t) 
		isComplete := true
		cfg.ProjectState.StrategicKickoffCompleted = &isComplete
		orc.assumeYes = true // To bypass branch prompt in daily, assuming it would be called

		// orchestrator's gitClient is nil from setupTestOrchestrator
		err := orc.ExecuteKickoff(ctx, false /*isStrategicFlag*/, "test-daily-branch") 
		
		require.Error(t, err) 
		assert.Contains(t, err.Error(), "git client not available for daily kickoff")
		assert.Contains(t, outBuf.String(), "Starting Daily Git Kickoff workflow") 
		assert.NotContains(t, outBuf.String(), "Strategic Project Kickoff - Prompt Generation")
	})
}

func TestMarkStrategicKickoffComplete(t *testing.T) {
	ctx := context.Background()
	orc, cfg, presenterOut, _, _, _, configFilePath := setupTestOrchestrator(t)

	initialKickoffCompleted := false
	if cfg.ProjectState.StrategicKickoffCompleted != nil {
		initialKickoffCompleted = *cfg.ProjectState.StrategicKickoffCompleted
	}
	assert.False(t, initialKickoffCompleted, "Kickoff should not be complete initially by default")

	err := orc.MarkStrategicKickoffComplete(ctx)
	require.NoError(t, err)

	require.NotNil(t, cfg.ProjectState.StrategicKickoffCompleted)
	assert.True(t, *cfg.ProjectState.StrategicKickoffCompleted, "StrategicKickoffCompleted should be true in memory")
	assert.NotEmpty(t, cfg.ProjectState.LastStrategicKickoffDate, "LastStrategicKickoffDate should be set")

	loadedCfg, err := config.LoadConfig(configFilePath)
	require.NoError(t, err)
	require.NotNil(t, loadedCfg)
	require.NotNil(t, loadedCfg.ProjectState.StrategicKickoffCompleted)
	assert.True(t, *loadedCfg.ProjectState.StrategicKickoffCompleted, "StrategicKickoffCompleted should be true in file")
	assert.Equal(t, cfg.ProjectState.LastStrategicKickoffDate, loadedCfg.ProjectState.LastStrategicKickoffDate)

	assert.Contains(t, presenterOut.String(), "Strategic kickoff has been marked as complete")
}

// TODO: Add tests for runCollaborationSetup (mock presenter, check config updates)
// TODO: Add tests for runInitialInfoGathering (mock presenter, check map result)
// TODO: Add tests for generateCollaborationPrefsYAML (check YAML string output)
// TODO: Add tests for generateMasterKickoffPromptText (check template execution, parameter substitution)
// TODO: Add extensive tests for executeDailyKickoff (this will require Orchestrator to use a GitClientInterface)

