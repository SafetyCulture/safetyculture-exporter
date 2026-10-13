# MCP Support — AI Agents & Interactive UIs (FG-9869)

> **Status:** Phases 1–4 implemented on branch `FG-9869-mcp-ai-agents` (draft PR #516).
> User guide: [connect-ai.md](connect-ai.md).

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
| `query` | Arbitrary read-only SQL `SELECT` | row-capped (default 200, max 1000), 30s timeout; visibility `[model, app]` |
| `sync_data` | SQLite export (same as `mitti-exporter sqlite`) | mutating; elicits incremental vs full where supported |
| `generate_report` | PDF/WORD inspection reports via `RunInspectionReports` | renders in report-viewer app |
| `read_media` | Inspection photos as MCP image content | path-jailed to media dir, 25MB cap |
| `read_export_file` | File bytes for the report viewer | **app-only** (`visibility: ["app"]`) |
| `sync_progress` | Live per-feed export progress | **app-only** |

Prompts: `weekly_safety_review`, `overdue_actions`, `site_scorecard(site)`.

### Query safety

Defence in depth on `query`:

1. DB opened with `mode=ro` — writes are impossible at the SQLite level.
2. Statement guard: single statement only, must start with `SELECT`/`WITH`/`EXPLAIN`,
   write keywords (`INSERT`, `DROP`, `PRAGMA`, `ATTACH`, …) rejected.
3. Row cap + `context.WithTimeout` bound cost.
4. Every result includes a text summary, so hosts without MCP Apps support
   degrade gracefully to plain text.

## MCP Apps (SEP-1865)

Three self-contained HTML resources (`text/html;profile=mcp-app`), embedded via
`go:embed`, zero external requests (empty CSP = maximally strict):

- `ui://mitti/table-viewer` — query results as an interactive table with
  client-side column sorting (zero model tokens for re-sorting)
- `ui://mitti/report-viewer` — generated reports with Download buttons
  (fetches bytes via the app-only `read_export_file` tool, saves via
  `ui/download-file`; text fallback lists the on-disk paths)
- `ui://mitti/sync-status` — live per-feed progress, polling the app-only
  `sync_progress` tool while `sync_data` runs

Tools declare `_meta.ui.resourceUri`, so supporting hosts render results in the
apps; others show the JSON text fallback. Views do the SEP-1865 handshake
(`ui/initialize` → `ui/notifications/initialized`), honour host light/dark
theming, and report `ui/notifications/size-changed`.

## Distribution

- **Desktop app (primary):** the Wails binary doubles as the server —
  `<binary> mcp` runs headless using the app's settings. "Connect AI" page
  writes the Claude Desktop config (merge-only, with backup). Token never
  leaves the app's config file.
- **CLI:** `mitti-exporter mcp`; falls back to the desktop app's settings file
  when no local config exists.
- **MCPB:** `scripts/pack-mcpb.sh <platform> <binary> <version>` packs a
  one-click `.mcpb` (binary server type, no secrets in the manifest); the
  release workflow attaches darwin/win32 bundles to each GitHub Release.

## Roadmap / later

- Sampling (SEP-1577): "summarise this table" inside an MCP App view
- Streamable HTTP transport for web connectors (only if demand; local stdio first)
- Prompt-injection guard: wrap user-authored record text (notes, comments) as
  untrusted data in tool outputs
- PII pseudonymisation option for agent-facing outputs
- MCPB signing (`mcpb sign`) once a signing identity is chosen

## Testing

```bash
cd cli
GOWORK=off go test ./internal/mcp/...

# Manual smoke test (needs a sqlite_export.db in ./export):
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smoke","version":"0"}}}' \
  | ./mitti-exporter mcp
```
