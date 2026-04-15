# Plan 02-02 Execution Summary

**Plan:** UI CI Workflow (`ci-ui.yml`)
**Status:** Complete
**Date:** 2026-04-15

## Tasks Executed

### Task 2.1 — Create ci-ui.yml workflow file

**Action:** Created `.github/workflows/ci-ui.yml` with three jobs.

**File created:** `.github/workflows/ci-ui.yml`

**Jobs:**
1. `unit-tests` — Runs on `ubuntu-latest`, runs `go test ./...` from `working-directory: ui` with Go workspace active (no `GOWORK=off`).
2. `build-smoke-test` — Runs on `macos-latest`, sets up pnpm v10 (before Node.js), Node.js 20 with pnpm cache, installs Wails v2.12.0, and runs `wails build -platform darwin/universal -clean`.
3. `govulncheck` — Runs on `ubuntu-latest`, installs and runs `govulncheck ./...` from `working-directory: ui` with workspace active.

**Trigger paths:** `ui/**`, `go.work`, `go.work.sum`, `.github/workflows/ci-ui.yml`

## Acceptance Criteria Results

| Criterion | Result |
|-----------|--------|
| File `.github/workflows/ci-ui.yml` exists | PASS |
| Contains `pull_request:` with `types: [opened, synchronize, reopened]` | PASS |
| Contains `paths:` with `'ui/**'` | PASS |
| Does NOT contain `GOWORK: off` | PASS |
| `pnpm/action-setup@v4` appears before `actions/setup-node@v4` | PASS |
| `version: 10` under pnpm/action-setup | PASS |
| `node-version: '20'` | PASS |
| `cache: 'pnpm'` | PASS |
| `cache-dependency-path: ui/frontend/pnpm-lock.yaml` | PASS |
| `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0` | PASS |
| `wails build -platform darwin/universal -clean` | PASS |
| `working-directory: ui` in test and build steps | PASS |
| `actions/checkout@v4` and `actions/setup-go@v5` | PASS |
| `govulncheck` job runs `govulncheck ./...` from `working-directory: ui` | PASS |
| `govulncheck` job does NOT contain `GOWORK: off` | PASS |

## Key Design Decisions

- Go workspace is kept active (no `GOWORK=off`) so the UI module's dependency on the local CLI module (`github.com/SafetyCulture/safetyculture-exporter`) resolves via `go.work`.
- Build smoke test targets `macos-latest` because Wails v2 requires native CGO and platform headers for `darwin/universal` fat binary compilation.
- pnpm setup (`pnpm/action-setup@v4`) intentionally precedes Node.js setup (`actions/setup-node@v4`) — this is required for pnpm cache registration in GitHub Actions.
- Wails version pinned to `v2.12.0` matching `ui/go.mod`.

## Commits

- `feat(02-02): create UI CI workflow (ci-ui.yml)` — task 2.1
- `docs(02-02): add execution summary for plan 02-02` — this commit
