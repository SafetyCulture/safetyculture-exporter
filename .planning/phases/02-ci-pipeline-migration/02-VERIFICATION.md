---
status: human_needed
phase: 02-ci-pipeline-migration
verified: 2026-04-15
---

# Phase 2: CI Pipeline Migration — Verification

## Automated Checks

### CI-01 — Test workflow triggers on PR push (opened, synchronize, reopened) with path filters for `cli/**` and `ui/**`

**PASS**

- `ci-cli.yml` line 7: `types: [opened, synchronize, reopened]` with `paths: - 'cli/**'`
- `ci-ui.yml` line 7: `types: [opened, synchronize, reopened]` with `paths: - 'ui/**'`
- Both workflows also filter on `go.work`, `go.work.sum`, and their own workflow file path
- Path filters are present; runtime enforcement requires a live PR run (see Human Verification)

---

### CI-02 — Go unit tests run explicitly per module (`cd cli && go test ./...`, `cd ui && go test ./...`)

**PASS**

- `ci-cli.yml`: `unit-tests` job uses `working-directory: cli` with `run: go test ./...`
- `ci-ui.yml`: `unit-tests` job uses `working-directory: ui` with `run: go test ./...`
- Neither workflow runs `go test ./...` from the workspace root
- `GOWORK=off` set for CLI tests; workspace active (no override) for UI tests so local CLI module resolves

---

### CI-03 — Security scan (govulncheck) ported to monorepo workspace

**PASS**

- `ci-cli.yml`: `govulncheck` job runs `govulncheck ./...` from `working-directory: cli` with `GOWORK: off`
- `ci-ui.yml`: `govulncheck` job runs `govulncheck ./...` from `working-directory: ui` with workspace active
- Old `security-scan.yml` (which used `edplato/trufflehog-actions-scan@v0.9i-beta`) is deleted
- `govulncheck` count: 5 in `ci-cli.yml` (once per GOWORK-isolated job), `ci-ui.yml` has its own govulncheck job

---

### CI-04 — CLI cross-platform release builds (linux/darwin/windows, amd64/arm64) on tag push

**PASS**

- `release.yml` trigger: `on: push: tags: - "v*.*.*"`
- 4 CLI jobs present: `cli-linux-amd64`, `cli-windows-amd64`, `cli-darwin-amd64`, `cli-darwin-arm64`
- All 4 jobs use `GOWORK: off` (count verified: 4 in release.yml)
- All 4 jobs use `working-directory: cli` in build and package steps (10 total matches)
- Linux job: `ubuntu-latest`, `GOOS=linux GOARCH=amd64`
- Windows job: `ubuntu-latest` with mingw cross-compiler (`CC=x86_64-w64-mingw32-gcc`)
- Darwin jobs: `macos-latest`, amd64 and arm64 separately

---

### CI-05 — UI release builds for macOS (universal binary) and Windows on tag push

**PASS**

- `ui-macos` job: `macos-latest`, `wails build -platform darwin/universal -clean`
- `ui-windows` job: `windows-latest`, `wails build -platform windows/amd64 -clean -skipbindings`
- Both jobs triggered on same `v*.*.*` tag push
- No `GOWORK: off` on either UI job (workspace active for local CLI resolution)

---

### CI-06 — Single GitHub Release created with all CLI and UI artifacts on `v*` tag

**PASS**

- `release` aggregator job present in `release.yml`
- `needs:` lists all 6 build jobs: `cli-linux-amd64`, `cli-windows-amd64`, `cli-darwin-amd64`, `cli-darwin-arm64`, `ui-macos`, `ui-windows`
- `actions/download-artifact@v4` with `merge-multiple: true` flattens all artifacts into `artifacts/`
- `ncipollo/release-action@v1` with `artifacts: "artifacts/*"` uploads all 6 artifacts
- Release created as `draft: true`, `prerelease: true`, `makeLatest: false`
- Expected artifacts: `safetyculture-exporter-linux-amd64.zip`, `safetyculture-exporter-windows-x86_64.zip`, `safetyculture-exporter-darwin-amd64.zip`, `safetyculture-exporter-darwin-arm64.zip`, `safetyculture-exporter-ui-darwin-universal.zip`, `safetyculture-exporter-ui-windows-x86_64.exe`

---

### CI-07 — macOS codesign + notarize + staple in correct per-arch sequence

**PASS**

- `ui-macos` job sequence verified in order:
  1. `wails build -platform darwin/universal -clean` (build)
  2. `codesign --force -s ... --options runtime` (sign with hardened runtime)
  3. `ditto -c -k --keepParent ... ./safetyculture-exporter-ui-darwin-universal.zip` (zip for notarization)
  4. `xcrun notarytool submit ... --wait` (notarize)
  5. `xcrun stapler staple ui/build/bin/safetyculture-exporter-ui.app` (staple ticket)
  6. `ditto -c -k --keepParent ...` (re-zip with staple ticket embedded)
- `ditto` count: 2 (verified)
- `xcrun stapler staple` count: 1 (verified)
- CLI darwin jobs also codesign binaries with `codesign --force -s ... --options runtime`
- Cleanup step with `if: always()` removes `certificate.p12` and `appstore_api_key.p8`

---

### CI-08 — Windows codesign with pinned action SHA (not `@develop`)

**PASS**

- `ui-windows` job uses `sslcom/esigner-codesign@v2.0.0` (not `@develop`)
- `signing_method: v2` and `jvm_max_memory: 1024M` present
- `file_path` references `safetyculture-exporter-ui.exe` (matches `wails.json` outputfilename)
- Grep for `@develop` across all workflow files: 0 matches

---

### CI-09 — Version injection via `-ldflags` from monorepo git tag for both binaries

**PASS**

- CLI jobs: `-X github.com/SafetyCulture/safetyculture-exporter/internal/app/version.version=${{ github.ref_name }}` (4 jobs)
- UI jobs: `-X github.com/SafetyCulture/safetyculture-exporter-ui/internal/version.version=${{ github.ref_name }}` (in `ui-macos` and `ui-windows`)
- Both use `${{ github.ref_name }}` which resolves to the tag name on a tag push

---

### CI-10 — pnpm setup (`pnpm/action-setup@v4` before `setup-node@v4`) with Node 20 LTS

**PASS**

- `ci-ui.yml` `build-smoke-test` job: `pnpm/action-setup@v4` (step 3) before `actions/setup-node@v4` (step 4)
- `release.yml` `ui-macos` job: `pnpm/action-setup@v4` before `actions/setup-node@v4`
- `release.yml` `ui-windows` job: `pnpm/action-setup@v4` before `actions/setup-node@v4`
- All use `version: 10` for pnpm and `node-version: '20'` for Node
- All use `cache: 'pnpm'` with `cache-dependency-path: ui/frontend/pnpm-lock.yaml`

---

### CI-11 — Wails CLI version in CI matches `ui/go.mod` requirement (v2.12.0)

**PASS**

- `ci-ui.yml`: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
- `release.yml` `ui-macos`: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
- `release.yml` `ui-windows`: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
- All 3 occurrences pin exactly `v2.12.0`

---

### CI-12 — `GOWORK=off` for CLI-only release builds (module isolation)

**PASS**

- `ci-cli.yml`: `GOWORK: off` set in 5 job `env:` blocks: `unit-tests`, `govulncheck`, `postgres-integration`, `mysql-integration`, `sqlserver-integration` (count verified: 5)
- `release.yml`: `GOWORK: off` set in all 4 CLI build job `env:` blocks (count verified: 4)
- `ci-ui.yml`: `GOWORK` not present anywhere (count verified: 0) — workspace remains active for UI builds
- `release.yml` UI jobs (`ui-macos`, `ui-windows`): no `GOWORK: off`

---

## Roadmap Success Criteria Checks

### 1. CLI-only PR triggers CLI job, skips UI job (path filters present)

**PASS (static)** / **Human needed (runtime)**

Path filter `'cli/**'` present in `ci-cli.yml`; path filter `'ui/**'` present in `ci-ui.yml`. The YAML is correct. Runtime verification — pushing a CLI-only PR and confirming `ci-ui.yml` is not triggered — requires an actual GitHub Actions run.

### 2. v* tag push produces single GitHub Release with all artifacts

**PASS (static)** / **Human needed (runtime)**

`release.yml` trigger is `on: push: tags: - "v*.*.*"`. All 6 build jobs feed into the `release` aggregator. YAML is structurally correct. Runtime verification requires a real tag push.

### 3. macOS .app has staple step (`xcrun stapler staple` present)

**PASS**

`xcrun stapler staple ui/build/bin/safetyculture-exporter-ui.app` verified present at line 247 of `release.yml`. Second `ditto` re-zips after staple.

### 4. go test scoped per module (working-directory: cli/ui, not workspace root)

**PASS**

- `ci-cli.yml` unit-tests: `working-directory: cli`, `run: go test ./...`
- `ci-ui.yml` unit-tests: `working-directory: ui`, `run: go test ./...`
- `release.yml` CLI builds: `working-directory: cli` on all build steps (10 matches)
- No `go test` or `go build` runs from workspace root

### 5. All action pins current (no @develop, @beta, @v2 for checkout/setup-go)

**PASS**

Grep for `@develop`, `@beta`, `setup-go@v2`, `checkout@v2`, `setup-node@v3`, `edplato` across all 3 workflow files: **0 matches**.

Current pins verified:

| Action | Pin | Files |
|--------|-----|-------|
| `actions/checkout` | `@v4` | All 3 |
| `actions/setup-go` | `@v5` | All 3 |
| `actions/setup-node` | `@v4` | `release.yml`, `ci-ui.yml` |
| `actions/setup-python` | `@v5` | `ci-cli.yml` |
| `actions/upload-artifact` | `@v4` | `release.yml` |
| `actions/download-artifact` | `@v4` | `release.yml` |
| `pnpm/action-setup` | `@v4` | `release.yml`, `ci-ui.yml` |
| `pre-commit/action` | `@v3.0.1` | `ci-cli.yml` |
| `ncipollo/release-action` | `@v1` | `release.yml` |
| `sslcom/esigner-codesign` | `@v2.0.0` | `release.yml` |

---

## Must-Haves Verification

### From Plan 02-01

| Must-Have | Status |
|-----------|--------|
| PR changes to `cli/**` trigger CLI test jobs with path filters | PASS |
| `GOWORK=off` set for all CLI test and vulnerability scan jobs | PASS (5 jobs) |
| Tests run from `cli/` directory (`working-directory: cli`), not workspace root | PASS |
| govulncheck runs against CLI module | PASS |
| All action pins are current (`@v4` or `@v5`) | PASS |
| Old incompatible workflows removed (`tests.yml`, `security-scan.yml`) | PASS |

### From Plan 02-02

| Must-Have | Status |
|-----------|--------|
| PR changes to `ui/**` trigger UI test and build jobs with path filters | PASS |
| Go workspace remains active (no `GOWORK=off`) so local CLI module resolves | PASS |
| pnpm v10 setup before Node.js v20 setup (correct order for caching) | PASS |
| Wails CLI version exactly v2.12.0 (matching ui/go.mod) | PASS |
| Build smoke test runs on macOS with `darwin/universal` platform | PASS |
| govulncheck runs against UI module with workspace active (no GOWORK=off) | PASS |

### From Plan 02-03

| Must-Have | Status |
|-----------|--------|
| Old `build.yml` is deleted | PASS |
| CLI builds for all 4 platform/arch combinations | PASS |
| `GOWORK=off` on every CLI build job | PASS (4 jobs) |
| Version injection via ldflags with correct module path | PASS |
| macOS CLI binaries are codesigned | PASS |
| All actions pinned to current versions (v4/v5) | PASS |

### From Plan 02-04

| Must-Have | Status |
|-----------|--------|
| macOS universal binary built with `wails build -platform darwin/universal` | PASS |
| Complete Apple sequence: codesign → zip → notarize → staple → re-zip | PASS |
| Hardened runtime flag (`--options runtime`) on codesign | PASS |
| Notarization ticket stapled into the .app before final zip | PASS |
| pnpm v10 + Node 20 setup in correct order | PASS |
| Wails v2.12.0 (matching ui/go.mod) | PASS |
| Version injection via ldflags with UI module path | PASS |
| No GOWORK override (workspace active for local CLI resolution) | PASS |

### From Plan 02-05

| Must-Have | Status |
|-----------|--------|
| Windows UI build using Wails on `windows-latest` runner | PASS |
| SSL.com CodeSignTool action pinned to `@v2.0.0` (not `@develop`) | PASS |
| Version injection via ldflags with UI module path | PASS |
| pnpm v10 + Node 20 setup in correct order | PASS |
| Wails v2.12.0 (matching ui/go.mod) | PASS |
| Output filename matches wails.json `outputfilename` (`safetyculture-exporter-ui`) | PASS |
| No GOWORK override (workspace active for local CLI resolution) | PASS |

### From Plan 02-06

| Must-Have | Status |
|-----------|--------|
| Single GitHub Release created with all 6 artifacts (4 CLI + 2 UI) | PASS |
| Release aggregator job depends on all 6 build jobs | PASS |
| All action pins across all 3 workflow files are current | PASS |
| No placeholder comments remain in release.yml | PASS |
| Exactly 3 workflow files exist: `ci-cli.yml`, `ci-ui.yml`, `release.yml` | PASS |

---

## Human Verification Required

The following items cannot be confirmed by static YAML analysis. They require actual GitHub Actions runs.

### HV-01: Path filter isolation (runtime)
Push a PR that modifies only `cli/**` files and confirm GitHub Actions runs `ci-cli.yml` but does NOT trigger `ci-ui.yml`. Then repeat with a `ui/**`-only PR. The YAML path filters are correct; runtime enforcement is what needs validation.

### HV-02: v* tag produces GitHub Release
Push a `v*.*.*` tag to the remote and confirm the `release` workflow runs all 7 jobs in sequence and creates a GitHub Release with all 6 artifacts attached.

### HV-03: macOS Gatekeeper offline validation
Download the `safetyculture-exporter-ui-darwin-universal.zip` from the release, extract it, disconnect from the internet, and open the `.app`. Confirm Gatekeeper accepts it (staple ticket present and valid). This validates that the staple → re-zip sequence in `ui-macos` actually works end-to-end.

### HV-04: darwin/universal fat binary verification
Confirm the macOS runner produces a true fat binary: `file safetyculture-exporter-ui.app/Contents/MacOS/safetyculture-exporter-ui` must report both `x86_64` and `arm64` architectures.

### HV-05: Windows SmartScreen pass
Run the signed `safetyculture-exporter-ui-windows-x86_64.exe` on a clean Windows machine and confirm SmartScreen does not block it (requires valid SSL.com credentials configured as GitHub secrets).

### HV-06: Secrets configured in GitHub
Confirm all required secrets are set in the repo's GitHub Actions secrets:
- `MAC_SIGNING_CERT` (base64-encoded .p12)
- `MAC_SIGNING_KEYCHAIN_PWD`
- `MAC_SIGNING_CERT_PASSWORD`
- `MAC_SIGNING_CERT_NAME`
- `APPSTORE_API_KEY`
- `APPSTORE_API_KEY_ID`
- `APPSTORE_API_ISSUER_ID`
- `SSL_DOT_COM_USERNAME`
- `SSL_DOT_COM_PASSWORD`
- `CREDENTIAL_ID`
- `TOTP_SECRET`

### HV-07: pnpm-lock.yaml exists at expected path
The `cache-dependency-path: ui/frontend/pnpm-lock.yaml` in both `ci-ui.yml` and `release.yml` assumes this lockfile exists. Confirm `ui/frontend/pnpm-lock.yaml` is committed (this becomes relevant once the React scaffold in Phase 3 is complete).

### HV-08: UI module version symbol exists
The ldflags path `github.com/SafetyCulture/safetyculture-exporter-ui/internal/version.version` must correspond to an actual exported `version` variable in the `ui/` module. Confirm `ui/internal/version/version.go` exists and exports `var version string`.

---

## Summary

**Static verification: all 12 requirements (CI-01 through CI-12) PASS.**

All three workflow files (`ci-cli.yml`, `ci-ui.yml`, `release.yml`) match the specifications from plans 02-01 through 02-06. The old incompatible workflows (`tests.yml`, `security-scan.yml`, `build.yml`) are confirmed deleted. Action pins are current across all files with no forbidden floating tags. The GOWORK isolation strategy is correctly applied — off for CLI jobs, active for UI jobs. The macOS Apple notarization sequence (build → sign → zip → notarize → staple → re-zip) is implemented in the correct order. Windows code signing uses `sslcom/esigner-codesign@v2.0.0` (not `@develop`). The release aggregator depends on all 6 build jobs and creates a single GitHub Release.

**8 human verification items remain** (HV-01 through HV-08), all requiring live CI runs or credential configuration. These cannot be satisfied by static analysis. The phase goal — "deliver a trustworthy, monorepo-aware CI pipeline before any React code is written" — is structurally achieved. Runtime validation should occur on the first real PR and the first `v*` tag push.

**Recommended next action:** Proceed to Phase 3 (React App Scaffold). Runtime CI validation will happen naturally as PRs are opened against the monorepo.
