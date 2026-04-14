# Architectural Review Report — Allele UI Re-Architecture

**Subject:** Allele UI Re-Architecture: Pro Financial Terminal
**Date:** 2026-04-13
**Mode:** Document Review
**Panel:** Context Master (Gemini 3 Pro) · The Architect (Claude Opus 4.6) · Security Sentinel (OpenAI o4) · Product Visionary (GPT-5.2) · Creative Strategist (GPT-5.3-Codex) · The Naysayer (Claude Opus 4.6)

---

## Final Recommendation: Request Changes

The shift to a "Pro Financial Terminal" aesthetic and modular architecture is strategically sound and will massively improve usability and credibility. However, the panel identified several **Critical** and **Major** risks—primarily that the frontend vision vastly outpaces the current backend API capabilities, and that expanding the current unauthenticated API poses severe security risks. 

The panel recommends proceeding with a tightly scoped MVP, but **only after** designing the required backend API contracts and addressing the security boundaries.

---

## Findings Summary

| Severity | Count |
|----------|-------|
| Critical | 3     |
| Major    | 9     |
| Minor    | 5     |
| Info     | 2     |

---

## Critical Issues (Must Address)

### RISK-001: Backend API Surface is Woefully Insufficient
- **Severity:** Critical
- **Source:** The Naysayer
- **Description:** The proposal treats the re-architecture as a frontend problem. In reality, the current backend exposes almost nothing the new UI needs (no endpoints for kernel state, strategy performance, Sortino ratios, P&L, risk gate status, etc.). The entire "Pro Terminal" vision requires building 10-15 new REST/WS endpoints.
- **Recommendation:** Before touching the UI, create a Backend API Design Document that inventories every data point the new UI panels need, maps each to an existing or new Go endpoint, and sequences the backend work as Phase 0.

### SEC-001 / RISK-002: Unauthenticated REST API & CORS Risks
- **Severity:** Critical
- **Source:** Security Sentinel, The Naysayer
- **Description:** The current backend has `Access-Control-Allow-Origin: '*'` and the REST endpoints (`/api/plugins`, `/api/plugins/config`) require zero authentication. Expanding this API to handle more configuration (including API keys) without auth is a severe access-control gap (A01). Anyone on the network could POST to the config endpoint and change trading parameters.
- **Recommendation:** Replace wildcard CORS with a specific origin allowlist. Extend the `auth_token` mechanism from the WebSocket endpoint to all REST endpoints. Implement strict server-side input validation for all plugin configurations.

### PROD-001: Scope Risk - Building a Full Terminal Before Proving Value
- **Severity:** Critical
- **Source:** Product Visionary
- **Description:** The proposal bundles multiple big bets (docking, schema wizards, API changes). If delivered as one epic, it risks a long cycle time without validated user impact, optimizing for "looks like a terminal" rather than "makes traders faster."
- **Recommendation:** Define a strict MVP: Docking layout + local persistence + schema-driven config for 1-2 critical plugins + 3 core panes (Firehose, Trace, Chart). Defer the full component catalog until the core interaction model is validated.

---

## Major Issues (Should Address)

### ARCH-005: WebSocket Protocol Lacks Typed Message Envelope
- **Severity:** Major
- **Source:** The Architect
- **Description:** The current WS broadcasts raw JSON payloads. The UI parses everything as a generic log. There is no type discrimination to route ticks to the Order Book vs. health events to the Causality Trace.
- **Recommendation:** Implement a typed WS envelope (`{"type": "tick", "payload": ...}`) in `broadcaster.go`. Implement a UI-side message bus so panels only re-render when their specific data types arrive.

### CREAT-002: Prevent Configuration from Hijacking Context
- **Severity:** Major
- **Source:** Creative Strategist, Product Visionary
- **Description:** An "in-your-face wizard" is correct *only* when a plugin is blocking core system function. If every missing optional field triggers a blocking flow, users lose trust.
- **Recommendation:** Implement a 3-tier setup UX: (1) Hard-block preflight only for critical missing config, (2) soft banner for degraded mode, (3) quiet settings drawer for optional tuning.

### RISK-003: Layout Engine Selection is a High-Risk Decision
- **Severity:** Major
- **Source:** The Naysayer, The Architect
- **Description:** The wrong docking library will require a complete rewrite or create permanent UX compromises. `React-Grid-Layout` is a grid, not a docking system. `GoldenLayout` has React integration risks.
- **Recommendation:** `FlexLayout-React` is the strongest candidate. Run a 2-day prototype spike to validate tabs, nesting, and performance before committing.

### RISK-004: No Specification for Degraded/Disconnected UI States
- **Severity:** Major
- **Source:** The Naysayer
- **Description:** The current app has a naive reconnect loop and doesn't clearly indicate stale data. A financial terminal showing stale data without a massive warning is a material trading risk.
- **Recommendation:** Define a formal UI State Machine (CONNECTING, LIVE, STALE, ENGINE_ERROR). The Firehose and Order Book must prominently display "STALE" if updates stop. Implement exponential backoff for WS reconnects.

### RISK-007: Unbounded Log Accumulation & Render Performance
- **Severity:** Major
- **Source:** The Naysayer
- **Description:** The current UI keeps 100 logs in React state and re-renders the whole list on every tick. At 50-100 msgs/sec, this will cause layout thrashing and frame drops.
- **Recommendation:** Replace the log array with a virtualized list (e.g., `@tanstack/react-virtual`). Batch WS messages on `requestAnimationFrame` boundaries rather than updating state per message.

### ARCH-002 / PROD-005: Plugin Schema API Extension
- **Severity:** Major
- **Source:** The Architect, Product Visionary
- **Description:** The backend `abi.Manifest` exists but lacks UI metadata (validation, grouping, defaults) to drive a setup wizard effectively. However, over-designing it too early will stall development.
- **Recommendation:** Extend `abi.ConfigField` with minimal fields: `Validation` (regex/type), `Options` (enums), and `Default`. Add a `GET /api/plugins/{name}/ready` endpoint so the UI knows if the wizard must block.

---

## Minor Suggestions (Nice to Have)

### ARCH-003: Dual-Layer State Persistence
- **Severity:** Minor
- **Source:** The Architect, Creative Strategist
- **Description:** Layout state should live in `localStorage` (it's device-specific), while plugin config stays in SQLite.
- **Recommendation:** Use `localStorage` for the MVP FlexLayout model. Later, add a backend `layout_presets` table for syncing layouts across machines.

### CREAT-006: Role-Based Workspace Presets
- **Severity:** Minor
- **Source:** Creative Strategist
- **Description:** High density can overwhelm users if not structured.
- **Recommendation:** Ship the UI with default presets (e.g., "Execution Mode", "Monitoring Mode") that users can easily switch between.

### ARCH-007: Strip Cyberpunk Aesthetic
- **Severity:** Minor
- **Source:** The Architect
- **Description:** Scanlines, glowing text, and pulsing animations detract from the professional feel and eat GPU cycles.
- **Recommendation:** Move to a muted, high-contrast palette (navy/blacks, crisp grays, strict green/red accents). Remove `index.css` scanlines.

---

## What Was Done Well

*   **Platform Thinking:** The core premise of mirroring the kernel's highly modular architecture in the UI is strategically correct and sets the stage for a powerful tool.
*   **Schema Foundation:** The existing `abi.Manifest` system in the backend is a fantastic foundation. The fact that WASM plugins self-declare their config requirements is architecturally elegant.
*   **Storage Paradigm:** Using SQLite for `plugin_config` rather than `.env` files is the right call for a dynamic, API-driven system.
*   **WebSocket Focus:** Using a single WebSocket rather than REST polling aligns perfectly with the real-time requirements of algorithmic trading.

---

## Action Items

- [ ] **Phase 0 (Backend):** Write an API contract document defining the new endpoints needed for the UI (Leaderboard, Health, Strategy metrics).
- [ ] **Phase 0 (Security):** Lock down the `/api/plugins` endpoints with CORS allowlists, the `auth_token` middleware, and strict input validation.
- [ ] **Phase 1 (Backend):** Implement the `WSMessage` typed envelope in `broadcaster.go`.
- [ ] **Phase 2 (Frontend Spike):** Build a 2-day throwaway prototype using `flexlayout-react` to validate docking and virtualization performance.
- [ ] **Phase 3 (MVP):** Build the core terminal shell, the non-blocking/blocking plugin wizard logic, and the local layout persistence.
- [ ] **Phase 4 (Aesthetic):** Apply the new professional color palette and remove animations/scanlines.

---

*Generated by DSS Architectural Review Panel · 2026-04-13*