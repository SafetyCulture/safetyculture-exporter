// Package mcpcmd provides the `mitti-exporter mcp` command, which runs an MCP
// (Model Context Protocol) server over stdio so AI agents (Claude Desktop,
// Claude Code, etc.) can sync and query the local export database.
package mcpcmd

import (
	"context"

	"github.com/SafetyCulture/mitti-exporter/cmd/mitti-exporter/cmd/export"
	util "github.com/SafetyCulture/mitti-exporter/cmd/mitti-exporter/cmd/utils"
	"github.com/SafetyCulture/mitti-exporter/internal/app/version"
	mittimcp "github.com/SafetyCulture/mitti-exporter/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Cmd is the `mitti-exporter mcp` command.
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Run an MCP server so AI agents can query your exported Mitti data",
		Long: `Run a Model Context Protocol (MCP) server over stdio.

Add it to Claude Desktop's config (claude_desktop_config.json):

  {
    "mcpServers": {
      "mitti": {
        "command": "/path/to/mitti-exporter",
        "args": ["mcp", "--config-path", "/path/to/mitti-exporter.yaml"]
      }
    }
  }

The server exposes tools for agents to sync data (sync_data) and run
read-only SQL queries (list_tables, describe_table, query) against the
local SQLite export database, plus MCP Apps (SEP-1865) UI resources for
interactive results in supporting hosts.`,
		// Stdio transport: never print anything to stdout.
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runMCP,
	}
}

func runMCP(cmd *cobra.Command, _ []string) error {
	v := viper.GetViper()
	exportPath := v.GetString("export.path")

	cfg := &mittimcp.Config{
		Name:       "mitti-exporter",
		Version:    version.GetVersion(),
		ExportPath: exportPath,
		Sync: func(ctx context.Context) error {
			// Build the exporter lazily per sync so config/token changes are
			// picked up and a failed sync doesn't poison the server.
			exp := export.NewMittiExporter(v)
			return exp.RunSQLite()
		},
	}

	if err := mittimcp.Run(cmd.Context(), cfg); err != nil && err != context.Canceled {
		util.Check(err, "MCP server stopped with an error")
	}
	return nil
}
