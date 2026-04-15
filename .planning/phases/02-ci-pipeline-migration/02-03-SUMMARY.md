# Plan 02-03 Execution Summary

**Plan:** 02-03 — Release Workflow: CLI Builds
**Status:** Complete
**Executed:** 2026-04-15

## Tasks Completed

### Task 3.1 — Delete old build.yml workflow
- Deleted `.github/workflows/build.yml`
- The old workflow used outdated action pins (`@v2`), had no `GOWORK=off` isolation, used an inefficient matrix strategy with `if:` guards, and lacked separate artifact naming
- Commit: `chore(02-03): delete old build.yml workflow`

### Task 3.2 — Create release.yml with CLI build jobs
- Created `.github/workflows/release.yml` with 4 CLI platform jobs: `cli-linux-amd64`, `cli-windows-amd64`, `cli-darwin-amd64`, `cli-darwin-arm64`
- Commit: `feat(02-03): create release.yml with CLI cross-platform build jobs`

## Acceptance Criteria Verification

| Criterion | Result |
|-----------|--------|
| `build.yml` does not exist | PASS |
| `release.yml` exists | PASS |
| Tag trigger `v*.*.*` present | PASS |
| 4 CLI jobs defined | PASS (8 matches — job ID + job name) |
| All 4 jobs have `GOWORK: off` | PASS (4 matches) |
| All build/package steps use `working-directory: cli` | PASS (10 matches) |
| All 4 jobs inject version via ldflags | PASS (4 matches) |
| Both darwin jobs have `codesign --force -s` step | PASS (2 matches) |
| Both darwin jobs import signing certificate | PASS (2 matches) |
| Windows job uses `CC=x86_64-w64-mingw32-gcc` | PASS |
| All actions pinned to v4/v5 | PASS (12 action references) |
| `GO_VERSION: '1.23'` at workflow level | PASS |

## Key Design Decisions

- Separate jobs per platform (vs matrix with `if:` guards) — cleaner logs, easier debugging
- `GOWORK: off` as job-level env ensures CLI module isolation from Wails CGO deps
- `working-directory: cli` on build/package steps so commands run from CLI module root
- macOS jobs run on `macos-latest` for native arm64 and amd64 signing support
- Linux and Windows use `ubuntu-latest`; Windows cross-compiles with mingw
- Version injection path confirmed: `github.com/SafetyCulture/safetyculture-exporter/internal/app/version.version`
- `THIRD_PARTY_NOTICES.md` referenced without path prefix (local to `cli/`); `README.md` referenced as `../README.md` (repo root)
- Placeholder comments at end of file for Plans 02-04, 02-05 (UI jobs), and 02-06 (release aggregator)

## Files Modified

- `.github/workflows/build.yml` — deleted
- `.github/workflows/release.yml` — created
