# Architectural Review Report — Allele Tri-Plugin Microkernel

**Subject:** Allele Architecture & Tri-Plugin Microkernel
**Date:** 2026-04-13
**Mode:** Conversation Context
**Panel:** Context Master (Gemini 3 Pro) · The Architect (Claude Opus 4.6) · Security Sentinel (OpenAI o4) · Product Visionary (GPT-5.2) · Creative Strategist (GPT-5.3-Codex) · The Optimizer (GPT-5.3-Codex) · The Naysayer (Claude Opus 4.6)

---

## Final Recommendation: Request Changes

The Allele Engine possesses an exceptional conceptual foundation. The Hexagonal Microkernel, the Tri-Plugin WASM isolation boundaries, the GA Fairness Sidelining model, and the IPC Watchdog are highly defensible and innovative patterns for algorithmic trading. However, the panel identified **Critical** implementation gaps between the documented architecture and the required execution realities. Specifically: there is currently no hardcoded capital enforcement (`RiskGate`) in the hot path, the `SQLite` Vault stores credentials in plaintext, the `Health Monitor` is disconnected from the main event loop, and the WASM ABI bridge lacks the serialization required to actually execute WASM strategies. These must be addressed before any live capital is deployed.

---

## Findings Summary

| Severity | Count |
|----------|-------|
| Critical | 5     |
| Major    | 5     |
| Minor    | 3     |
| Info     | 2     |

---

## Critical Issues (Must Address)

### C-01: Missing Capital Enforcement (RiskGate) in Hot Path
- **Severity:** Critical
- **Source:** The Architect, The Naysayer
- **Description:** The system mandates a strict $100 capital cap and asymmetric risk bans, but there is no code path between `strategy.Evaluate()` and `exchange.SubmitOrder()` that enforces this. A buggy WASM strategy could return an unbounded `Size` value, draining the entire account in one API call.
- **Recommendation:** Implement a synchronous `RiskGate` layer in the Kernel that intercepts `[]TradeSignal` before submission. It must enforce: (1) $100 global cap, (2) per-strategy pot limits, (3) reject sizes exceeding available capital, and (4) mandatory stop-loss attachment.

### C-02: Plaintext Secrets in SQLite Vault (A02 Crypto)
- **Severity:** Critical
- **Source:** Security Sentinel, The Naysayer, The Architect
- **Description:** The Vault currently stores API keys and OAuth tokens as plaintext `TEXT` columns in SQLite. While WASM sandboxing prevents strategies from reading them, any host-level access (or backup script) exposes all exchange credentials.
- **Recommendation:** Encrypt credentials at rest in the SQLite database using Go's `crypto/aes` (AES-GCM) with a key derived from a secure environment variable or OS keyring. Ensure the `.db` file has `0600` permissions.

### C-03: Disconnected Health Monitor & Sidelining
- **Severity:** Critical
- **Source:** The Architect, The Naysayer
- **Description:** The `health.Monitor` and `arena.SidelineOrganism()` logic is perfectly tested but entirely disconnected from the Kernel's event loop. If a Sensor crashes, the Circuit Breaker trips, but the Kernel continues feeding stale data to the Strategy and executing its trades.
- **Recommendation:** Inject the `Monitor` and `Arena` into the Kernel. Wrap `strategy.Evaluate()` in a check for `monitor.IsStrategyPaused(id)`. If paused, skip evaluation and ensure the GA is actively discounting that time.

### C-04: No WebSocket Reconnection Logic (Silent Death)
- **Severity:** Critical
- **Source:** The Architect, The Naysayer
- **Description:** Polymarket/Exchange WebSockets drop frequently. Currently, a read error causes the listen goroutine to silently exit. Ticks stop flowing, the engine hangs, and the system relies on the 30-second IPC Watchdog timeout to kill and restart the container, leaving the system blind during volatile market moves.
- **Recommendation:** Implement an exponential backoff reconnect loop (1s, 2s, 4s... max 30s) inside the Exchange adapters. On reconnect, automatically resubscribe to all active markets.

### C-05: Incomplete WASM ABI Bridge
- **Severity:** Critical
- **Source:** The Architect, Security Sentinel
- **Description:** The codebase has two disconnected type systems: `core.MarketState` (used by the engine) and `abi.MarketState` (intended for WASM). There is no serialization bridge (e.g., JSON/FlatBuffers) or shared memory allocation implemented in the wazero loader to actually pass data into a WASM `Evaluate()` function.
- **Recommendation:** Define a canonical serialization format. Implement a `WasmStrategy` adapter that wraps `loader.WasmModule`, handles shared memory allocation, serializes the `MarketState` into the guest, calls the exported WASM function, and deserializes the returned `[]TradeSignal`.

---

## Major Issues (Should Address)

### M-01: SQLite Vault Contention on Hot Path
- **Severity:** Major
- **Source:** The Optimizer
- **Description:** Fetching API keys from SQLite at the exact millisecond of order execution will cause lock contention and latency spikes when multiple strategies fire simultaneously.
- **Recommendation:** Implement an in-memory credential cache with a short TTL, backed by SQLite for durability/rotation. Ensure SQLite is in WAL mode.

### M-02: Synchronous Kernel Execution Pipeline
- **Severity:** Major
- **Source:** The Architect, The Optimizer
- **Description:** The Kernel evaluates strategies and submits orders sequentially. A slow HTTP call to an exchange blocks the next strategy from executing and blocks the ingestion of the next WebSocket tick.
- **Recommendation:** Fan-out `strategy.Evaluate()` into parallel goroutines. Collect all signals, pass them through a serial `RiskGate`, and dispatch approved orders via a bounded worker pool.

### M-03: Unauthenticated UI WebSocket Broadcaster (A01 Access Control)
- **Severity:** Major
- **Source:** The Architect, The Naysayer
- **Description:** The Dashboard WebSocket (`:8081`) accepts all origins. Anyone on the network can connect and view live trading signals, organism fitness, and portfolio positions.
- **Recommendation:** Bind the broadcaster to `127.0.0.1:8081` (localhost only) or implement a shared-secret token in the WebSocket upgrade handshake.

### M-04: Strategy Authoring & Proof Workflow is Undefined
- **Severity:** Major
- **Source:** Product Visionary
- **Description:** The strict "No Black Boxes" rule requires mathematical proofs, but there is no formalized workflow tying a Markdown proof to a compiled WASM artifact. This creates a massive developer experience bottleneck.
- **Recommendation:** Create a CLI command (`allele prove <strategy>`) that links a standardized proof template to a strategy's build hash, ensuring auditable lineage.

### M-05: Sidelining UX is Opaque
- **Severity:** Major
- **Source:** Product Visionary, Creative Strategist
- **Description:** Auto-halting strategies for fairness is mathematically sound but frustrating for users who just see their bot "stop working" without explanation.
- **Recommendation:** Build a "Causality Trace" UI in the Global Radar that explicitly links: `Sensor Failure` -> `Circuit Breaker Tripped` -> `Strategy Halted` -> `Fairness Sidelining Active`.

---

## Minor Suggestions (Nice to Have)

### MIN-01: Narrower WASI FS Mounts
- **Severity:** Minor
- **Source:** The Architect
- **Description:** The WASI configuration mounts the entire plugin directory. It should ideally only mount the specific `config.yaml` file to adhere to least-privilege principles.
- **Recommendation:** Restrict the `wazero.FSConfig` to a targeted virtual overlay.

### MIN-02: Enrich MarketState with Orderbook Data
- **Severity:** Minor
- **Source:** The Architect, The Naysayer
- **Description:** `MarketState` currently only tracks the last-seen price. Arbitrage strategies need the Best Bid and Best Ask to calculate executable (post-slippage) profitability.
- **Recommendation:** Update the ABI to include `BestBid` and `BestAsk` instead of a singular `Price`.

### MIN-03: Freshness Metadata on Sensor Signals
- **Severity:** Minor
- **Source:** Creative Strategist
- **Description:** LLM or Twitter sensor data may be seconds/minutes old, while Exchange ticks are milliseconds old. Strategies need to know the age of the data.
- **Recommendation:** Attach a `Timestamp` or `AgeMs` field to all external signals injected into the `MarketState`.

---

## Informational Notes

### INF-01: Tri-Plugin Overhead vs Capital Cap
- **Source:** Creative Strategist
- **Description:** The heavy orchestration of WASM sandboxes might be over-engineered for a $100 fund. However, since the goal is a scalable, language-agnostic evolutionary arena, the architectural foundation is appropriate for the long-term vision.

### INF-02: $100 Cap & GA Signal Starvation
- **Source:** Product Visionary
- **Description:** Live trading with $100 will produce sparse fills and slow feedback loops for the Genetic Algorithm. The system will rely heavily on deterministic historical replay to generate enough fitness signal for evolution.

---

## What Was Done Well

- **Hexagonal Architecture:** The strict separation of `IExchange`, `IStrategy`, and `IEngine` is genuinely implemented and provides excellent isolation.
- **WASM for Sandboxing:** Using wazero (pure Go) to enforce zero-network math-only execution for strategies perfectly aligns with the "No Black Boxes" security requirement.
- **GA Fairness Sidelining:** The time-discounting fitness calculation that pauses during infrastructure outages is an innovative and highly effective solution to live evolutionary optimization.
- **Zero-CGO Toolchain:** Relying on `modernc.org/sqlite` and `wazero` guarantees a portable, cross-platform binary with minimal deployment friction.
- **Decoupled IPC Watchdog:** The out-of-process watchdog with conditional heartbeats is a mature and robust resilience pattern.

---

## Blind Voting Results (If Applicable)

None needed. The panel reached consensus on the necessity of the RiskGate, Vault encryption, and Kernel wiring.

---

## Action Items

- [ ] Implement `RiskGate` middleware to enforce $100 cap and position sizing.
- [ ] Add AES-GCM encryption at rest to the SQLite `Vault`.
- [ ] Wire `health.Monitor` and `arena.Arena` into the `Kernel` event loop.
- [ ] Implement exponential backoff WebSocket reconnection in Exchange adapters.
- [ ] Build the `WasmStrategy` Go wrapper to handle memory allocation and ABI serialization for wazero.
- [ ] Implement an in-memory cache for Vault credentials.
- [ ] Fan-out `strategy.Evaluate()` calls into a parallel goroutine pool.

---

*Generated by DSS Architectural Review Panel · 2026-04-13*
