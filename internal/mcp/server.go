package mcp

import (
	"context"
	"fmt"

	"github.com/contextvibes/cli/internal/globals"
	"github.com/mark3labs/mcp-go/mcp"    //nolint:depguard
	"github.com/mark3labs/mcp-go/server" //nolint:depguard
)

// Server wraps the MCP server instance.
type Server struct {
	mcpServer *server.MCPServer
}

// NewServer initializes the MCP server with standard configuration.
func NewServer() *Server {
	s := server.NewMCPServer(
		"ContextVibes",
		globals.AppVersion,
		server.WithLogging(),
	)

	return &Server{
		mcpServer: s,
	}
}

// StartStdio starts the server on Standard Input/Output.
// This blocks until the connection closes.
func (s *Server) StartStdio() error {
	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("mcp: failed to start stdio server: %w", err)
	}

	return nil
}

// AddTool registers a tool with the server.
func (s *Server) AddTool(tool mcp.Tool, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	s.mcpServer.AddTool(tool, handler)
}
