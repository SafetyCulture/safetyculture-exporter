# SafetyCulture Exporter Monorepo

## What This Is

A monorepo consolidation of the SafetyCulture Exporter CLI and its desktop UI wrapper. The CLI (Go) and UI (Wails + React/Tailwind/shadcn) live side-by-side using Go workspaces, with the UI referencing the CLI as a local module instead of published build artifacts. A unified CI pipeline publishes both CLI binaries and desktop app bundles as GitHub Release assets.

## Core Value

One repository, one release process — the UI always builds against the latest CLI code without waiting for artifact publishing.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Monorepo structure using Go workspaces (`go.work`) with `cli/` and `ui/` directories
- [ ] CLI code moved into `cli/` with preserved module path (`github.com/SafetyCulture/safetyculture-exporter`)
- [ ] UI code brought into `ui/` with `replace` directive pointing at local `cli/` module
- [ ] Frontend rewritten from Svelte to React + Tailwind CSS + shadcn/ui (pnpm, Vite)
- [ ] All existing UI screens replicated: Welcome, Init/Config, DataSetFilter, TemplateFilter, ExportStatus, Update
- [ ] Wails integration updated to work with new React frontend
- [ ] Consolidated GitHub Actions CI — build, test, security scan for both CLI and UI
- [ ] Release workflow publishes CLI binaries + UI desktop app (macOS/Windows) as GitHub Release assets
- [ ] All existing CLI tests pass from new location
- [ ] UI builds and runs against local CLI module

### Out of Scope

- New UI features or UX redesign — replicate existing screens only
- Mobile app or web-only deployment — Wails desktop only
- Changing the CLI's public API or module path
- Nx or monorepo tooling beyond Go workspaces — keep it simple
- Preserving UI repo git history — fresh copy, old history stays in original repo

## Context

- **CLI repo:** `safetyculture-exporter` — Go CLI tool for exporting SafetyCulture data. Has `cmd/`, `internal/`, `pkg/` structure with existing CI (build, test, security scan workflows).
- **UI repo:** `safetyculture-exporter-ui` — Wails v2 desktop app wrapping the CLI. Currently imports CLI as a versioned Go dependency. Frontend is Svelte 3 with Vite, ~8 routes/pages, custom components.
- **Reference project:** `/Users/neo.sheikh/Documents/web-platform` — internal project using pnpm, Tailwind, shadcn patterns to follow for the frontend rewrite.
- **Current pain:** Two repos means the UI must wait for CLI artifact publishing before it can use new CLI changes. Maintenance overhead of keeping dependency versions in sync.

## Constraints

- **Go workspaces**: Must use `go.work` for local module references — no vendoring hacks
- **Module path**: CLI module path (`github.com/SafetyCulture/safetyculture-exporter`) must not change — external consumers depend on it
- **Frontend stack**: React + Tailwind CSS + shadcn/ui with pnpm (not npm) and Vite
- **Wails v2**: UI must remain a Wails v2 desktop application
- **Release**: Single GitHub Release with both CLI and UI artifacts

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go workspaces over single go.mod | Keeps CLI and UI as separate modules with independent dependency trees; CLI module path preserved | — Pending |
| Fresh copy of UI code (no git history) | Simpler than subtree merge; old history remains in original repo if needed | — Pending |
| React + Tailwind + shadcn over Svelte | Aligns with internal web-platform patterns; better component ecosystem | — Pending |
| pnpm over npm | Matches web-platform reference project; faster, stricter dependency resolution | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-15 after initialization*
