# Plan 2 Summary: Copy UI Source into ui/

**Phase:** 01 - Monorepo Foundation
**Plan:** 2 of 3
**Status:** Complete

## What Was Done

- Copied Wails UI application from `/Users/neo.sheikh/Documents/safetyculture-exporter-ui/` into `ui/` at the monorepo root.
- Files copied: `go.mod`, `go.sum`, `main.go`, `app.go`, `wails.json`, `build/`, `internal/`, and `frontend/` (excluding `node_modules/`, `dist/`, `wailsjs/`).
- Created `ui/frontend/dist/.gitkeep` to satisfy the `go:embed` directive.
- Updated `ui/wails.json`:
  - `outputfilename` changed from `safetyculture-exporter` → `safetyculture-exporter-ui`
  - Added `wailsjsdir`: `./frontend/src/wailsjs`
  - Changed `frontend:install` from `npm install` → `pnpm install`
  - Changed `frontend:build` / `frontend:dev:watcher` from `npm run` → `pnpm run`

## Fix Applied

The original `//go:embed frontend/dist` directive excludes dotfiles, so `.gitkeep` was invisible to the embedder. Updated to `//go:embed all:frontend/dist` to include dotfiles.

## Verification Results

| Check | Result |
|-------|--------|
| `ui/go.mod` module = `github.com/SafetyCulture/safetyculture-exporter-ui` | PASS |
| `ui/main.go` exists | PASS |
| `ui/app.go` exists | PASS |
| `wails.json` `outputfilename` = `safetyculture-exporter-ui` | PASS |
| `wails.json` `wailsjsdir` = `./frontend/src/wailsjs` | PASS |
| `ui/frontend/dist/.gitkeep` exists | PASS |
| No `ui/frontend/node_modules/` | PASS |
| `go build ./...` in `ui/` | PASS (warnings only; CLI module resolves from current go.sum) |

## Notes

- `go build` succeeds because `ui/go.sum` already pins the CLI module version. Plan 3 (go.work) will switch resolution to the local `cli/` source.
- Git history from the source repo was not preserved (per plan spec).
