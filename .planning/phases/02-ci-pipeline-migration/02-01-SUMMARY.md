---
plan: 02-01
title: CLI CI Workflow
status: complete
completed_at: "2026-04-15"
---

# Plan 02-01 Execution Summary

## Objective

Create `.github/workflows/ci-cli.yml` for monorepo-aware CLI testing with `GOWORK=off` isolation, and remove the two old workflows that were incompatible with the monorepo structure.

## Tasks Completed

### Task 1.1 — Create ci-cli.yml
- Created `.github/workflows/ci-cli.yml` with 6 jobs:
  - `unit-tests`: runs `go test ./...` in `cli/` with `GOWORK=off`
  - `pre-commit`: runs pre-commit linting with Python setup
  - `govulncheck`: installs and runs govulncheck against the CLI module with `GOWORK=off`
  - `postgres-integration`: PostgreSQL integration tests with `GOWORK=off`
  - `mysql-integration`: MySQL 8 integration tests with `GOWORK=off`
  - `sqlserver-integration`: SQL Server integration tests with `GOWORK=off`
- Path filters restrict trigger to `cli/**`, `go.work`, `go.work.sum`, `.github/workflows/ci-cli.yml`, `.pre-commit-config.yaml`
- All action pins updated to current versions (`actions/checkout@v4`, `actions/setup-go@v5`, `actions/setup-python@v5`)
- `GOWORK: off` set in 5 job `env:` blocks (all except pre-commit)

### Task 1.2 — Delete tests.yml
- Removed `.github/workflows/tests.yml`
- Old workflow was incompatible: ran `go test ./...` from workspace root (no `go.mod`), had no path filters, used outdated `@v2` action versions, only triggered on `opened` PRs

### Task 1.3 — Delete security-scan.yml
- Removed `.github/workflows/security-scan.yml`
- Old workflow used a beta-pinned action (`edplato/trufflehog-actions-scan@v0.9i-beta`) with no path filters
- Trufflehog config files (`.github/trufflehog/`) retained for potential future reuse
- Go vulnerability scanning is now provided by the `govulncheck` job in `ci-cli.yml`

## Verification Results

| Check | Result |
|-------|--------|
| `ci-cli.yml` exists | PASS |
| `GOWORK: off` appears 5 times | PASS (count=5) |
| `working-directory: cli` in all test steps | PASS (5 matches) |
| `govulncheck ./...` present | PASS |
| `go-version: '1.23'` used throughout | PASS |
| All 3 DB service jobs present | PASS |
| `tests.yml` deleted | PASS |
| `security-scan.yml` deleted | PASS |
| `.github/workflows/` contains only `build.yml` and `ci-cli.yml` | PASS |

## Files Changed

- `.github/workflows/ci-cli.yml` — created (176 lines)
- `.github/workflows/tests.yml` — deleted
- `.github/workflows/security-scan.yml` — deleted

## Commits

1. `feat(02-01): create ci-cli.yml workflow with GOWORK=off isolation`
2. `chore(02-01): delete tests.yml workflow incompatible with monorepo`
3. `chore(02-01): delete security-scan.yml workflow with deprecated action`
