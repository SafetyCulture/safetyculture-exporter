---
plan: 02-04
title: "Release workflow — UI macOS build with signing/notarization/staple"
status: complete
executed_by: agent-a49a5f0d
date: 2026-04-15
---

# Plan 02-04 Execution Summary

## What Was Done

Added the `ui-macos` job to `.github/workflows/release.yml`, implementing the full macOS build and notarization pipeline for the Wails desktop UI.

## Task Outcomes

### Task 4.1: Add ui-macos job to release.yml (complete)

- Replaced the placeholder comment `# UI jobs are added by Plans 02-04 and 02-05` with the full `ui-macos` job definition.
- The existing CLI jobs (`cli-linux-amd64`, `cli-windows-amd64`, `cli-darwin-amd64`, `cli-darwin-arm64`) were preserved untouched.

## Key Implementation Details

**Apple signing/notarization sequence (correct order):**
1. Build macOS universal binary (`wails build -platform darwin/universal -clean`)
2. Codesign the `.app` bundle with hardened runtime (`--options runtime`)
3. Zip the signed `.app` for submission (`ditto`)
4. Submit to Apple notarization service (`xcrun notarytool submit --wait`)
5. Staple the notarization ticket into the `.app` (`xcrun stapler staple`)
6. Re-zip with stapled ticket for final artifact (`ditto` again)

**Critical configuration:**
- `GOWORK` not overridden — workspace active so `ui/go.mod` resolves the local CLI module
- `pnpm/action-setup@v4` placed before `actions/setup-node@v4` (required for pnpm cache to work)
- Wails pinned to `v2.12.0` matching `ui/go.mod`
- ldflags use the UI module path: `github.com/SafetyCulture/safetyculture-exporter-ui/internal/version.version`
- `.app` name derived from `wails.json` `outputfilename: safetyculture-exporter-ui`
- Cleanup step uses `if: always()` to remove `certificate.p12` and `appstore_api_key.p8` even on failure

## Acceptance Criteria Verification

- [x] `ui-macos` job present in release.yml
- [x] No `GOWORK: off` in `ui-macos` job (only on CLI jobs)
- [x] `pnpm/action-setup@v4` before `actions/setup-node@v4`
- [x] `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
- [x] `wails build -platform darwin/universal -clean`
- [x] ldflags with UI module path and `${{ github.ref_name }}`
- [x] `codesign --force -s` with `--options runtime`
- [x] `xcrun notarytool submit` with `--wait`
- [x] `xcrun stapler staple ui/build/bin/safetyculture-exporter-ui.app`
- [x] Second `ditto` command after staple step (re-zip)
- [x] `safetyculture-exporter-ui.app` used in all paths (not `safetyculture-exporter.app`)
- [x] Cleanup step with `if: always()`
- [x] Comment `# UI jobs are added by Plans 02-04 and 02-05` removed

## Commits

- `feat(02-04): add ui-macos job to release workflow with signing/notarization/staple`
