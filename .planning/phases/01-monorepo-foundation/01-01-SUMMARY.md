# Phase 1, Plan 1: Move CLI Source into cli/ Subdirectory

## Summary

Used `git mv` to relocate all CLI source files from the repository root into a new `cli/` subdirectory. The Go module path (`github.com/SafetyCulture/safetyculture-exporter`) was not changed.

## Files Moved

- `go.mod` → `cli/go.mod`
- `go.sum` → `cli/go.sum`
- `cmd/` → `cli/cmd/`
- `internal/` → `cli/internal/`
- `pkg/` → `cli/pkg/`
- `Makefile` → `cli/Makefile`
- `Dockerfile` → `cli/Dockerfile`
- `docker-compose.yml` → `cli/docker-compose.yml`
- `docker-compose-local-volume.yml` → `cli/docker-compose-local-volume.yml`
- `CONTRIBUTING.md` → `cli/CONTRIBUTING.md`
- `THIRD_PARTY_NOTICES.md` → `cli/THIRD_PARTY_NOTICES.md`
- `docs/` → `cli/docs/`

## Files NOT Moved (remain at repo root)

- `.git/`, `.github/`, `LICENSE`, `README.md`, `.gitignore`, `.planning/`, `CLAUDE.md`

## Verification Results

| Check | Result |
|-------|--------|
| `cli/go.mod` module path | `module github.com/SafetyCulture/safetyculture-exporter` — unchanged |
| Import paths containing `/cli/` | Zero found |
| `cd cli && go build ./...` | Exit 0 (warnings only from upstream dep) |
| `cd cli && go test ./...` | All packages pass, zero FAIL |
| `cd cli && GOWORK=off go build ./...` | Exit 0 (isolated build succeeds) |

## Acceptance Criteria Status

- [x] `cli/go.mod` exists with `module github.com/SafetyCulture/safetyculture-exporter`
- [x] `cli/cmd/safetyculture-exporter/` exists
- [x] `cli/pkg/` and `cli/internal/` exist
- [x] `go.mod` no longer exists at repo root
- [x] `cmd/` no longer exists at repo root
- [x] `cd cli && go build ./...` exits 0
- [x] `cd cli && go test ./...` exits 0 with no FAIL
- [x] `cd cli && GOWORK=off go build ./...` exits 0
- [x] Zero import paths contain `/cli/`
