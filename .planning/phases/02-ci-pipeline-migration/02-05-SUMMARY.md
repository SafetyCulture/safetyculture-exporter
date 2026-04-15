# Plan 02-05 Execution Summary

**Plan:** 02-05 — Release Workflow — UI Windows Build with Signing
**Date:** 2026-04-15
**Status:** Complete

## Tasks Executed

### Task 5.1 — Add ui-windows job to release.yml

Added the `ui-windows` job to `.github/workflows/release.yml` after the existing `ui-macos` job.

**Key implementation details:**
- Runs on `windows-latest` runner (native Windows build — no cross-compilation)
- No `GOWORK: off` override — workspace stays active so `ui/` resolves `cli/` locally
- pnpm/action-setup@v4 installed BEFORE actions/setup-node@v4 (required order for caching)
- `choco install mingw` provides CGO toolchain for Wails + sqlite dependencies
- Build step uses `shell: pwsh` with PowerShell environment variable manipulation for GOPATH
- `wails build -platform windows/amd64 -clean -skipbindings` with version ldflags injection
- Signing via `sslcom/esigner-codesign@v2.0.0` (pinned, not `@develop`)
- `file_path` uses `safetyculture-exporter-ui.exe` matching `wails.json` `outputfilename`
- `signing_method: v2` and `jvm_max_memory: 1024M` for SSL.com CodeSignTool
- Final artifact uploaded as `safetyculture-exporter-ui-windows-x86_64.exe`

## Verification Results

| Check | Result |
|-------|--------|
| `ui-windows` job present | PASS |
| Runs on `windows-latest` | PASS |
| No `GOWORK: off` | PASS |
| `pnpm/action-setup@v4` before `setup-node@v4` | PASS |
| `wails@v2.12.0` install | PASS |
| `wails build -platform windows/amd64 -clean -skipbindings` | PASS |
| Version ldflags with UI module path | PASS |
| `esigner-codesign@v2.0.0` (not `@develop`) | PASS |
| `safetyculture-exporter-ui.exe` in file_path | PASS |
| `choco install mingw` | PASS |
| `shell: pwsh` on build step | PASS |
| `signing_method: v2` | PASS |
| Aggregator comment preserved | PASS |

## Files Modified

- `.github/workflows/release.yml` — added `ui-windows` job (64 lines inserted)

## Commits

- `feat(02-05): add ui-windows job with SSL.com code signing to release workflow`
