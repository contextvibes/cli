package mcp

import (
	"bytes"
	"context"
	"fmt"

	"github.com/contextvibes/cli/cmd/product/quality"
	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/mark3labs/mcp-go/mcp" //nolint:depguard
)

// RegisterQualityTool adds the 'run-quality-checks' tool to the server.
func (s *Server) RegisterQualityTool() {
	tool := mcp.NewTool("run-quality-checks",
		mcp.WithDescription("Run code quality checks (linting, vet, security). Returns the report even if checks fail."),
		mcp.WithString("mode",
			mcp.Description("The quality check mode. Defaults to 'local' to use the project's own configuration."),
			// CHANGE: Set default to "local"
			mcp.DefaultString("local"),
			mcp.Enum("local", "essential", "strict", "style", "complexity", "security"),
		),
	)

	s.AddTool(tool, s.handleQualityCheck)
}

func (s *Server) handleQualityCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var mode string
	if argsMap, ok := request.Params.Arguments.(map[string]any); ok {
		mode, _ = argsMap["mode"].(string)
	}

	// CHANGE: Fallback to "local" if not specified
	if mode == "" {
		mode = "local"
	}

	globals.AppLogger.Info("MCP: Running quality checks", "mode", mode)

	// 1. Capture Output
	var outputBuf bytes.Buffer

	presenter := ui.NewPresenter(&outputBuf, &outputBuf)

	// 2. Run the Logic
	results, err := quality.RunQualityChecks(ctx, presenter, mode, nil)

	// 3. Construct Response
	response := fmt.Sprintf("## Execution Log\n\n%s\n", outputBuf.String())

	if err != nil {
		response += fmt.Sprintf("\n\n❌ **Outcome:** Pipeline Failed (%v)\n", err)
		if len(results) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("System Error: %v", err)), nil
		}
	} else {
		response += "\n\n✅ **Outcome:** All Checks Passed\n"
	}

	return mcp.NewToolResultText(response), nil
}
