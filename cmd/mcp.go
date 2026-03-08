package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/springcli/springcli/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the Model Context Protocol (MCP) Server",
	Long: `Starts SpringCLI as an MCP stdio server.
This allows AI assistants like Claude Desktop, Cursor, and Gemini
to programmatically control the CLI and manage Spring projects.
Output colorization and interactive interactive prompts are disabled automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Prevent any other tools or loggers from polluting stdout,
		// because stdio MCP strictly relies on clean JSON-RPC across stdin/stdout.
		// For an MCP server, we must disable human-centric interactive UI elements.

		server := mcp.NewServer()
		return server.Start(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
