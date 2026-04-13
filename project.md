# project.md — BeadBoard Driver Session Cache

This file is maintained by agents. A new agent reads this first.
If the Environment Status table shows all `pass`, skip straight to Step 2 of the runbook.
Only re-run a check if its row says `fail` or `unknown`, or if you hit an actual error.

---

## Environment Status Cache

Last updated: 2026-04-12 by `bb-trading-architect`

| Component | Status | Version / Detail | Verified |
|-----------|--------|-----------------|---------|
| `bd` on PATH | `pass` | /home/kenton/.local/bin/bd | 2026-04-12 |
| `bb` on PATH | `pass` | fnm_multishells.../bb | 2026-04-12 |
| `.beads` db exists | `pass` | `bd init` success | 2026-04-12 |
| `mail.delegate` configured | `pass` | `/home/kenton/.agents/skills/beadboard-driver/scripts/bb-mail-shim.mjs` | 2026-04-12 |
| `session-preflight` | `pass` | ok: true | 2026-04-12 |
| `bb agent` registered | `pass` | `BB_AGENT=trading-architect` | 2026-04-12 |
| Tests last run | `unknown` | | |

**Status values:** `pass` · `fail` · `unknown` · `skip` (not applicable to this project)

**Rule:** If every row is `pass` → skip Step 1 entirely and go straight to Step 2.
If any row is `fail` or `unknown` → run only that check, update this table, continue.

---

## Project Identity

- Project name: allele
- Repository root: /home/kenton/Documents/allele
- Primary language/runtime: Go
- Primary package manager: go mod

## Tooling Baseline

- `bd` installed and on PATH: yes
- `bb` installed and on PATH: yes
- Shell/platform: WSL2/Linux

## BeadBoard/Communication Setup

- `.beads` database: created on 2026-04-12 via `bd init`
- Mail delegate: `bd config set mail.delegate "node /home/kenton/.agents/skills/beadboard-driver/scripts/bb-mail-shim.mjs"` — configured 2026-04-12
- Agent identity policy: `export BB_AGENT=<role-name>` (set fresh each session in Step 2)
- `session-preflight` last pass: 2026-04-12

## Agent State + Heartbeat Policy

- Agent bead naming: `bb-<role-name>` (e.g. `bb-trading-architect`)
- Required state transitions: `spawning → running → working → stuck/done/stopped`
- Heartbeat: LLM agents heartbeat at turn start + before long commands; daemon agents every 5 min

## Command Baseline

- Install: `go mod tidy`
- Build: `go build ./cmd/allele`
- Typecheck: `go build ./...`
- Lint: `go vet ./...`
- Test: `go test ./...`

## Session Log (append-only)

Each agent appends one line when they update this file:

| Date | Agent | What changed |
|------|-------|-------------|
| 2026-04-12 | `bb-trading-architect` | Initial project.md created, all checks pass |
