// Package mcpcmd provides the `mitti-exporter mcp` command, which runs an MCP
// (Model Context Protocol) server over stdio so AI agents (Claude Desktop,
// Claude Code, etc.) can sync and query the local export database.
package mcpcmd

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/SafetyCulture/mitti-exporter/cmd/mitti-exporter/cmd/export"
	util "github.com/SafetyCulture/mitti-exporter/cmd/mitti-exporter/cmd/utils"
	"github.com/SafetyCulture/mitti-exporter/internal/app/version"
	exporterAPI "github.com/SafetyCulture/mitti-exporter/pkg/api"
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

If no local config file is found, the command falls back to the desktop
app's settings (so UI-only installs work too). The API token is read from
the config file only — never from MCP host configs.

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
	fallbackToUIConfig(v)

	exportPath := v.GetString("export.path")
	if !filepath.IsAbs(exportPath) {
		// Anchor relative export paths to the config file's directory — MCP
		// hosts spawn us with an unpredictable working directory.
		if cfgFile := v.ConfigFileUsed(); cfgFile != "" {
			exportPath = filepath.Join(filepath.Dir(cfgFile), exportPath)
		}
	}
	mediaPath := v.GetString("export.media_path")
	if mediaPath != "" && !filepath.IsAbs(mediaPath) {
		if cfgFile := v.ConfigFileUsed(); cfgFile != "" {
			mediaPath = filepath.Join(filepath.Dir(cfgFile), mediaPath)
		}
	}

	// makeExporter builds a fresh exporter per call so config/token changes
	// are picked up and a failed run doesn't poison the server.
	makeExporter := func(sync *mittimcp.SyncOptions, reportFormats []string) (*exporterAPI.MittiExporter, error) {
		cm, err := exporterAPI.NewConfigurationManagerFromFile("", v.ConfigFileUsed())
		if err != nil {
			return nil, err
		}
		export.MapViperConfigToExporterConfiguration(v, cm.Configuration)
		cm.ApplySafetyGuards()
		if sync != nil {
			if sync.Full {
				cm.Configuration.Export.Incremental = false
			}
			if len(sync.Tables) > 0 {
				cm.Configuration.Export.Tables = sync.Tables
			}
			if len(sync.TemplateIDs) > 0 {
				cm.Configuration.Export.TemplateIds = sync.TemplateIDs
			}
		}
		if len(reportFormats) > 0 {
			cm.Configuration.Report.Format = reportFormats
		}
		ver := exporterAPI.AppVersion{
			IntegrationID:      version.GetIntegrationID(),
			IntegrationVersion: version.GetVersion(),
		}
		return exporterAPI.NewMittiExporter(cm.Configuration, &ver)
	}

	cfg := &mittimcp.Config{
		Name:       "mitti-exporter",
		Version:    version.GetVersion(),
		ExportPath: exportPath,
		MediaPath:  mediaPath,
		Sync: func(ctx context.Context, opts mittimcp.SyncOptions) error {
			exp, err := makeExporter(&opts, nil)
			if err != nil {
				return err
			}
			return exp.RunSQLite()
		},
		GenerateReports: func(ctx context.Context, formats []string) ([]string, error) {
			for i, f := range formats {
				formats[i] = strings.ToUpper(f)
			}
			exp, err := makeExporter(nil, formats)
			if err != nil {
				return nil, err
			}
			if err := exp.RunInspectionReports(); err != nil {
				return nil, err
			}
			return mittimcp.ListReportFiles(exportPath)
		},
	}

	if err := mittimcp.Run(cmd.Context(), cfg); err != nil && err != context.Canceled {
		util.Check(err, "MCP server stopped with an error")
	}
	return nil
}

// fallbackToUIConfig points viper at the desktop app's settings file when no
// local/flagged config file exists, so the standalone CLI and the desktop app
// share configuration. Explicit --config-path always wins.
func fallbackToUIConfig(v *viper.Viper) {
	used := v.ConfigFileUsed()
	if used != "" {
		if _, err := os.Stat(used); err == nil {
			return // real config found
		}
	}

	candidate := uiSettingsConfigPath()
	if candidate == "" {
		return
	}
	if _, err := os.Stat(candidate); err != nil {
		return
	}

	v.SetConfigFile(candidate)
	_ = v.ReadInConfig()
}

// uiSettingsConfigPath mirrors the desktop app's settings location
// (see ui/app.go GetSettingDirectoryPath).
func uiSettingsConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		for _, dir := range []string{"mitti-exporter", "safetyculture-exporter"} {
			p := filepath.Join(homeDir, "Library", "Application Support", dir, "mitti-exporter.yaml")
			if _, err := os.Stat(p); err == nil {
				return p
			}
			legacy := filepath.Join(homeDir, "Library", "Application Support", dir, "safetyculture-exporter.yaml")
			if _, err := os.Stat(legacy); err == nil {
				return legacy
			}
		}
	}
	return ""
}
