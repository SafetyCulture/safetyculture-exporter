package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	osRuntime "runtime"
	"strings"

	"github.com/SafetyCulture/mitti-exporter-ui/internal/version"
	exporterAPI "github.com/SafetyCulture/mitti-exporter/pkg/api"
	extLogger "github.com/SafetyCulture/mitti-exporter/pkg/logger"
	mittimcp "github.com/SafetyCulture/mitti-exporter/pkg/mcp"
)

// This file implements the app's dual-mode: the same desktop binary can run
// headless as an MCP server (`Mitti Exporter.app/.../mitti-exporter mcp`), so
// users who only ever install the UI can still connect AI agents (Claude
// Desktop etc.) without a separate CLI download. The API token stays in the
// app's own config file — nothing sensitive is written to MCP host configs.

// runMCPHeadless starts the MCP server over stdio using the app's settings
// and exits. Called from main() before Wails starts; never opens a window.
func runMCPHeadless() {
	settingsDir, err := CreateSettingsDirectory()
	if err != nil {
		fatalMCP("failed to get settings directory", err)
	}
	configFile := resolveConfigFileName(settingsDir)
	if !checkForConfigFile(settingsDir, configFile) {
		fatalMCP("no configuration found", fmt.Errorf("open the Mitti Exporter app once and enter your API token first (expected config at %s)", filepath.Join(settingsDir, configFile)))
	}

	cm, err := exporterAPI.NewConfigurationManagerFromFile(settingsDir, configFile)
	if err != nil {
		fatalMCP("failed to load configuration", err)
	}
	cm.ApplySafetyGuards()

	exportPath := cm.Configuration.Export.Path
	if !filepath.IsAbs(exportPath) {
		exportPath = filepath.Join(settingsDir, exportPath)
	}
	mediaPath := cm.Configuration.Export.MediaPath
	if mediaPath != "" && !filepath.IsAbs(mediaPath) {
		mediaPath = filepath.Join(settingsDir, mediaPath)
	}

	// makeExporter rebuilds config + exporter per call so token/config changes
	// are picked up without restarting the MCP server.
	makeExporter := func(sync *mittimcp.SyncOptions, reportFormats []string) (*exporterAPI.MittiExporter, error) {
		fresh, err := exporterAPI.NewConfigurationManagerFromFile(settingsDir, configFile)
		if err != nil {
			return nil, err
		}
		fresh.ApplySafetyGuards()
		if sync != nil {
			if sync.Full {
				fresh.Configuration.Export.Incremental = false
			}
			if len(sync.Tables) > 0 {
				fresh.Configuration.Export.Tables = sync.Tables
			}
			if len(sync.TemplateIDs) > 0 {
				fresh.Configuration.Export.TemplateIds = sync.TemplateIDs
			}
		}
		if len(reportFormats) > 0 {
			for i, f := range reportFormats {
				reportFormats[i] = strings.ToUpper(f)
			}
			fresh.Configuration.Report.Format = reportFormats
		}
		ver := exporterAPI.AppVersion{
			IntegrationID:      version.GetIntegrationID(),
			IntegrationVersion: version.GetVersion(),
		}
		return exporterAPI.NewMittiExporter(fresh.Configuration, &ver)
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

	// Logs go to stderr + file; stdout is reserved for JSON-RPC.
	extLogger.GetLogger()
	if err := mittimcp.Run(context.Background(), cfg); err != nil && err != context.Canceled {
		fatalMCP("MCP server stopped with an error", err)
	}
	os.Exit(0)
}

func fatalMCP(msg string, err error) {
	fmt.Fprintf(os.Stderr, "mitti-exporter mcp: %s: %s\n", msg, err.Error())
	os.Exit(1)
}

// --- "Connect AI" support (bound to the Wails frontend) ---

// MCPConnectInfo describes everything the Connect-AI page needs.
type MCPConnectInfo struct {
	BinaryPath        string `json:"binary_path"`
	ConfigSnippet     string `json:"config_snippet"`
	ClaudeConfigPath  string `json:"claude_config_path"`
	ClaudeConfigured  bool   `json:"claude_configured"`
	CursorConfigPath  string `json:"cursor_config_path"`
	CursorConfigured  bool   `json:"cursor_configured"`
	ClaudeCodeCommand string `json:"claude_code_command"`
	HasToken          bool   `json:"has_token"`
	HasDatabase       bool   `json:"has_database"`
	ExportPath        string `json:"export_path"`
}

// claudeDesktopConfigPath returns the Claude Desktop config file location.
func claudeDesktopConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch osRuntime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	default:
		return filepath.Join(homeDir, ".config", "claude", "claude_desktop_config.json"), nil
	}
}

// cursorConfigPath returns the Cursor MCP config file location.
func cursorConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cursor", "mcp.json"), nil
}

// mcpServerEntry builds the mcpServers.mitti entry pointing at this binary.
func mcpServerEntry(binaryPath string) map[string]any {
	return map[string]any{
		"command": binaryPath,
		"args":    []string{"mcp"},
	}
}

// resolvedBinaryPath returns the absolute, symlink-resolved path of this binary.
func resolvedBinaryPath() (string, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(binaryPath); err == nil {
		binaryPath = resolved
	}
	return binaryPath, nil
}

// GetMCPConnectInfo returns the binary path, ready-to-paste snippets per
// client, and status checks for the Connect-AI page.
func (a *App) GetMCPConnectInfo() (*MCPConnectInfo, error) {
	binaryPath, err := resolvedBinaryPath()
	if err != nil {
		return nil, err
	}

	snippet, err := json.MarshalIndent(map[string]any{
		"mcpServers": map[string]any{"mitti": mcpServerEntry(binaryPath)},
	}, "", "  ")
	if err != nil {
		return nil, err
	}

	claudePath, _ := claudeDesktopConfigPath()
	cursorPath, _ := cursorConfigPath()

	exportPath := a.cm.Configuration.Export.Path
	settingsDir, _ := GetSettingDirectoryPath()
	if !filepath.IsAbs(exportPath) {
		exportPath = filepath.Join(settingsDir, exportPath)
	}
	_, dbErr := os.Stat(filepath.Join(exportPath, "sqlite_export.db"))

	return &MCPConnectInfo{
		BinaryPath:        binaryPath,
		ConfigSnippet:     string(snippet),
		ClaudeConfigPath:  claudePath,
		ClaudeConfigured:  mcpConfigAlreadyWritten(claudePath, binaryPath),
		CursorConfigPath:  cursorPath,
		CursorConfigured:  mcpConfigAlreadyWritten(cursorPath, binaryPath),
		ClaudeCodeCommand: fmt.Sprintf("claude mcp add mitti -- \"%s\" mcp", binaryPath),
		HasToken:          a.cm.Configuration.AccessToken != "",
		HasDatabase:       dbErr == nil,
		ExportPath:        exportPath,
	}, nil
}

func mcpConfigAlreadyWritten(hostConfigPath, binaryPath string) bool {
	data, err := os.ReadFile(hostConfigPath)
	if err != nil {
		return false
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}
	servers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		return false
	}
	entry, ok := servers["mitti"].(map[string]any)
	if !ok {
		return false
	}
	return entry["command"] == binaryPath
}

// WriteMCPConfig merges the mitti server entry into an MCP client's config
// file, preserving any other servers. Backs up the existing file first.
// Never writes secrets — only the binary path and args.
// Supported clients: "claude" (Claude Desktop), "cursor".
func (a *App) WriteMCPConfig(client string) error {
	binaryPath, err := resolvedBinaryPath()
	if err != nil {
		return err
	}

	var hostConfigPath string
	switch client {
	case "claude":
		hostConfigPath, err = claudeDesktopConfigPath()
	case "cursor":
		hostConfigPath, err = cursorConfigPath()
	default:
		return fmt.Errorf("unsupported client %q — copy the snippet manually", client)
	}
	if err != nil {
		return err
	}

	cfg := map[string]any{}
	if data, err := os.ReadFile(hostConfigPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("existing config at %s is not valid JSON — fix or remove it first", hostConfigPath)
		}
		// Back up before modifying.
		if err := os.WriteFile(hostConfigPath+".bak", data, 0600); err != nil {
			return fmt.Errorf("failed to back up existing config: %w", err)
		}
	}

	servers, _ := cfg["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers["mitti"] = mcpServerEntry(binaryPath)
	cfg["mcpServers"] = servers

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(hostConfigPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(hostConfigPath, out, 0600)
}
