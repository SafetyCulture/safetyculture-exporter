package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	osRuntime "runtime"

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

	cfg := &mittimcp.Config{
		Name:       "mitti-exporter",
		Version:    version.GetVersion(),
		ExportPath: exportPath,
		Sync: func(ctx context.Context) error {
			ver := exporterAPI.AppVersion{
				IntegrationID:      version.GetIntegrationID(),
				IntegrationVersion: version.GetVersion(),
			}
			// Rebuild per sync so token/config changes are picked up.
			fresh, err := exporterAPI.NewConfigurationManagerFromFile(settingsDir, configFile)
			if err != nil {
				return err
			}
			fresh.ApplySafetyGuards()
			exp, err := exporterAPI.NewMittiExporter(fresh.Configuration, &ver)
			if err != nil {
				return err
			}
			return exp.RunSQLite()
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

// --- "Connect to Claude" support (bound to the Wails frontend) ---

// MCPConnectInfo describes everything the Connect-to-Claude page needs.
type MCPConnectInfo struct {
	BinaryPath    string `json:"binary_path"`
	HostConfigPath string `json:"host_config_path"`
	ConfigSnippet string `json:"config_snippet"`
	ConfigWritten bool   `json:"config_written"`
	HasToken      bool   `json:"has_token"`
	HasDatabase   bool   `json:"has_database"`
	ExportPath    string `json:"export_path"`
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

// mcpServerEntry builds the mcpServers.mitti entry pointing at this binary.
func mcpServerEntry(binaryPath string) map[string]any {
	return map[string]any{
		"command": binaryPath,
		"args":    []string{"mcp"},
	}
}

// GetMCPConnectInfo returns the binary path, a ready-to-paste config snippet,
// and status checks for the Connect-to-Claude page.
func (a *App) GetMCPConnectInfo() (*MCPConnectInfo, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	// Resolve symlinks so the config points at a stable path.
	if resolved, err := filepath.EvalSymlinks(binaryPath); err == nil {
		binaryPath = resolved
	}

	hostConfigPath, err := claudeDesktopConfigPath()
	if err != nil {
		return nil, err
	}

	snippet, err := json.MarshalIndent(map[string]any{
		"mcpServers": map[string]any{"mitti": mcpServerEntry(binaryPath)},
	}, "", "  ")
	if err != nil {
		return nil, err
	}

	exportPath := a.cm.Configuration.Export.Path
	settingsDir, _ := GetSettingDirectoryPath()
	if !filepath.IsAbs(exportPath) {
		exportPath = filepath.Join(settingsDir, exportPath)
	}
	_, dbErr := os.Stat(filepath.Join(exportPath, "sqlite_export.db"))

	return &MCPConnectInfo{
		BinaryPath:     binaryPath,
		HostConfigPath: hostConfigPath,
		ConfigSnippet:  string(snippet),
		ConfigWritten:  mcpConfigAlreadyWritten(hostConfigPath, binaryPath),
		HasToken:       a.cm.Configuration.AccessToken != "",
		HasDatabase:    dbErr == nil,
		ExportPath:     exportPath,
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

// WriteMCPConfig merges the mitti server entry into the Claude Desktop config,
// preserving any other servers the user has configured. Backs up the existing
// file first. Never writes secrets — only the binary path and args.
func (a *App) WriteMCPConfig() error {
	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(binaryPath); err == nil {
		binaryPath = resolved
	}

	hostConfigPath, err := claudeDesktopConfigPath()
	if err != nil {
		return err
	}

	cfg := map[string]any{}
	if data, err := os.ReadFile(hostConfigPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("existing Claude config at %s is not valid JSON — fix or remove it first", hostConfigPath)
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
