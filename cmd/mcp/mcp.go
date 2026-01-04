// Package mcp provides the command to start the Model Context Protocol server.
package mcp

import (
	"fmt"

	"github.com/contextvibes/cli/internal/mcp"
	"github.com/spf13/cobra"
)

// NewMcpCmd creates and configures the `mcp` command.
func NewMcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Starts the ContextVibes Model Context Protocol (MCP) server.",
		Long: `Starts a JSON-RPC 2.0 server over Stdio. 
This allows AI agents to directly access ContextVibes tools.`,
		Example:           `  contextvibes mcp`,
		GroupID:           "core",
		RunE:              runMcp,
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,

		// Boilerplate
		Aliases:                    []string{},
		SuggestFor:                 []string{},
		ValidArgs:                  []string{},
		ValidArgsFunction:          nil,
		Args:                       nil,
		ArgAliases:                 []string{},
		BashCompletionFunction:     "",
		Deprecated:                 "",
		Annotations:                nil,
		Version:                    "",
		PersistentPreRun:           nil,
		PersistentPreRunE:          nil,
		PreRun:                     nil,
		PreRunE:                    nil,
		Run:                        nil,
		PostRun:                    nil,
		PostRunE:                   nil,
		PersistentPostRun:          nil,
		PersistentPostRunE:         nil,
		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
		CompletionOptions:          cobra.CompletionOptions{},
		TraverseChildren:           false,
		Hidden:                     false,
		DisableFlagParsing:         false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}

	return cmd
}

func runMcp(_ *cobra.Command, _ []string) error {
	// 1. Initialize Server
	server := mcp.NewServer()

	// 2. Register Tools
	server.RegisterQualityTool()
	// Future: server.RegisterGitTools()
	// Future: server.RegisterProjectTools()

	// 3. Start (Blocking)
	if err := server.StartStdio(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	return nil
}
