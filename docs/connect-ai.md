# Connect an AI assistant to Mitti Exporter

Mitti Exporter can run as a local [MCP](https://modelcontextprotocol.io) server, so AI
assistants like Claude can sync and query your exported data — inspections, templates,
actions, issues, schedules, sites, assets and more — to answer questions and build reports.

**Your API token never leaves the exporter.** The MCP server reads it from the exporter's
own config file. Nothing sensitive is written to Claude's config.

## Option 1 — Desktop app (easiest)

1. Open the Mitti Exporter app and make sure your API token is set.
2. Run one **SQLite** export (or skip — Claude can trigger a sync itself).
3. Go to **Export configuration → Connect AI**.
4. Click **Connect to Claude**, then restart Claude Desktop.

## Option 2 — One-click bundle (.mcpb)

1. Download `mitti-exporter-darwin.mcpb` (macOS) or `mitti-exporter-win32.mcpb` (Windows)
   from the [latest release](https://github.com/SafetyCulture/mitti-exporter/releases).
2. Double-click it (or drag it onto Claude Desktop → Settings → Extensions).
3. The bundle uses the exporter's config file for your token — run
   `mitti-exporter configure` once first if you haven't set one up.

## Option 3 — Manual config

Add to Claude Desktop's `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mitti": {
      "command": "/path/to/mitti-exporter",
      "args": ["mcp", "--config-path", "/path/to/mitti-exporter.yaml"]
    }
  }
}
```

For Claude Code: `claude mcp add mitti -- /path/to/mitti-exporter mcp --config-path /path/to/mitti-exporter.yaml`

## What you can ask

- *"Sync my latest data, then show inspections that failed last month by site."*
- *"Which actions are overdue, grouped by assignee?"*
- *"Generate a PDF report for last week's inspections."* (download button appears in chat)
- *"Build a scorecard for Eastgate Yard."* (or use the built-in prompts:
  `weekly_safety_review`, `overdue_actions`, `site_scorecard`)

## Safety

- **Read-only by design.** Queries run against a *local copy* of your data. The database
  is opened read-only and only single `SELECT` statements are accepted — the assistant
  cannot change your Mitti data through this server.
- **Token stays local.** It's read from the exporter config file only.
- **Data stays on your machine.** Sync downloads via the same API the exporter already
  uses; queries never leave your computer. (What the AI assistant does with the answers
  is governed by that assistant's own privacy policy.)

## Troubleshooting

| Symptom | Fix |
|---|---|
| Claude says no database found | Run a SQLite export once, or ask Claude to "sync my data" |
| Server fails to start | Check the config path exists; run `mitti-exporter mcp` in a terminal to see the error on stderr |
| Token expired | Update it in the app (or config file) — no need to touch Claude's config |
| Reports/photos missing | Enable media export (`export.media: true`) and report formats in your config |
