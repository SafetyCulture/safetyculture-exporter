# MCP Support — AI Agents & Interactive UIs (FG-9869)

> **Status:** Spike / draft. Phase 1 implemented; Phases 2–3 proposed.

Mitti Exporter can run as an [MCP](https://modelcontextprotocol.io) server so AI agents
(Claude Desktop, Claude Code, other MCP hosts) can sync and query your exported data
locally. It also implements [SEP-1865 (MCP Apps)](https://modelcontextprotocol.io/seps/1865-mcp-apps-interactive-user-interfaces-for-mcp)
so supporting hosts render query results as an interactive UI instead of raw JSON.

## Quick start

```bash
# 1. Configure the exporter as usual (access token etc.)
mitti-exporter configure

# 2. Point your MCP host at the server. Claude Desktop (claude_desktop_config.json):
{
  "mcpServers": {
    "mitti": {
      "command": "/path/to/mitti-exporter",
      "args": ["mcp", "--config-path", "/path/to/mitti-exporter.yaml"]
    }
  }
}
```

Then ask Claude things like *"how many inspections failed last month by site?"* or
*"sync my latest data and summarise overdue actions"*.

## Architecture

```
Claude / MCP host
      │  stdio (JSON-RPC)
      ▼
mitti-exporter mcp                     cli/internal/mcp/
      │  sync_data ──────────────────▶ pkg/api MittiExporter.RunSQLite()
      │  list_tables / describe_table / query (read-only)
      ▼
<export-path>/sqlite_export.db         (existing `sqlite` export destination)
```

- **Transport:** stdio only (local-first; no open ports, no auth surface).
  Streamable HTTP for hosted/web connectors is a possible Phase 3.
- **Stdout hygiene:** the CLI version banner moves to stderr for `mcp`; zap logs
  already go to stderr + file. JSON-RPC owns stdout.
- **SDK:** `github.com/modelcontextprotocol/go-sdk` (official).

## Tools (search + execute style)

Instead of one tool per table (21+ tools, heavy on model context), the server
exposes a small discovery → execute surface:

| Tool | Purpose | Notes |
|---|---|---|
| `list_tables` | All tables with column/row counts | read-only, cheap manifest |
| `describe_table` | Schema + 5 sample rows for one table | read-only |
| `query` | Arbitrary read-only SQL `SELECT` | row-capped (default 200, max 1000), 30s timeout |
| `sync_data` | Incremental SQLite export (same as `mitti-exporter sqlite`) | mutating |

### Query safety

Defence in depth on `query`:

1. DB opened with `mode=ro` — writes are impossible at the SQLite level.
2. Statement guard: single statement only, must start with `SELECT`/`WITH`/`EXPLAIN`,
   write keywords (`INSERT`, `DROP`, `PRAGMA`, `ATTACH`, …) rejected.
3. Row cap + `context.WithTimeout` bound cost.
4. Every result includes a text summary, so hosts without MCP Apps support
   degrade gracefully to plain text.

## MCP Apps (SEP-1865)

- `ui://mitti/table-viewer` — self-contained HTML (`text/html;profile=mcp-app`),
  embedded via `go:embed`, zero external requests (empty CSP = maximally strict).
- The `query` tool declares `_meta.ui.resourceUri`, so supporting hosts render
  results in the interactive table; others show the JSON text fallback.
- The view does the SEP-1865 handshake (`ui/initialize` →
  `ui/notifications/initialized`), renders `ui/notifications/tool-result`
  structured content, reports `ui/notifications/size-changed`, and honours host
  light/dark theming via the standardized CSS variables.

## Roadmap

- **Phase 1 (this spike):** stdio server, 4 tools, table-viewer app, tests.
- **Phase 2:** report viewer app (`generate_report` tool → PDF/Word via
  `RunInspectionReports`, preview + `ui/download-file` button); sync progress app
  (`ui://mitti/sync-status` fed by `GetExporterStatus`); app-only tools
  (`visibility: ["app"]`) for UI-side sorting/pagination without model tokens.
- **Phase 3:** streamable HTTP transport for Claude.ai web connectors; one-click
  "Connect to Claude" page in the desktop UI that writes the host config; Windows
  install path handling.

## Testing

```bash
cd cli
GOWORK=off go test ./internal/mcp/...

# Manual smoke test (needs a sqlite_export.db in ./export):
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
  | ./mitti-exporter mcp
```
