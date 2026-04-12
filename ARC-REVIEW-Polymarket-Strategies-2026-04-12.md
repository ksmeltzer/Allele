# Architectural Review Report — Polymarket Algorithmic Trading Strategies

**Subject:** Polymarket Algorithmic Trading: Beyond Simple Arbitrage
**Date:** 2026-04-12
**Mode:** Document Review
**Panel:** Context Master (Gemini 3 Pro) · The Architect (Claude Opus 4.6) · Security Sentinel (OpenAI o4) · Product Visionary (GPT-5.2) · Creative Strategist (GPT-5.3-Codex) · The Optimizer (GPT-5.3-Codex) · The Naysayer (Claude Opus 4.6)

---

## Final Recommendation: Approve with Conditions (Prototype Spike Required)

The panel agrees that the technical constraints (Go, event-driven CLOB state manager, mathematical +EV focus) are highly sound. However, the panel unanimously warns that before complex strategies can be pursued, critical missing foundational pieces must be proven: specifically, the availability of historical tick data and the realism of the execution simulator. 

The panel has identified several highly novel, overlooked prediction market strategies (most notably **Bayesian Consistency Arbitrage** and **Resolution Timing Arbitrage**) but mandates a "Phase 0" smoke test first.

---

## Findings Summary

| Severity | Count |
|----------|-------|
| Critical | 2     |
| Major    | 6     |
| Minor    | 3     |
| Info     | 3     |

---

## Critical Issues (Must Address)

### RISK-001: No Historical Data Source Identified
- **Severity:** Critical
- **Source:** The Naysayer, Product Visionary
- **Description:** The entire architecture is predicated on backtesting using historical tick/order book data, yet Polymarket does not provide a public historical tick data API—only real-time WebSockets. If you cannot get historical data, the backtesting premise collapses.
- **Recommendation:** **PROTOTYPE SPIKE REQUIRED.** Before writing strategy logic, validate if historical data can be sourced. If not, immediately build a WebSocket data-recorder to start accumulating your own proprietary historical tick dataset.

### PROD-001: Backtesting Without Queue/Fill Realism Creates Fake Edge
- **Severity:** Critical
- **Source:** Creative Strategist, The Naysayer, Product Visionary
- **Description:** If the backtester assumes mid-price or top-of-book fills without modeling the queue and liquidity depth, every strategy will look profitable. In thin prediction markets, market impact and slippage will destroy paper PnL.
- **Recommendation:** The execution simulator must be a first-class component. Implement a fill model that walks the order book, incorporates Polymarket fees, and simulates partial fills and "leg risk" (when one side of an arbitrage fills but the other doesn't).

---

## Major Issues (Should Address)

### ARCH-004: UMA Oracle Resolution Inconsistency
- **Severity:** Major
- **Source:** The Architect, The Naysayer, Security Sentinel
- **Description:** In cross-market arbitrage, you hold opposing positions. Polymarket relies on UMA's optimistic oracle. Disputes can lock your capital for weeks. Worse, two logically linked markets could be resolved *inconsistently* by different proposers, turning a risk-free arbitrage into a total loss.
- **Recommendation:** Build an OracleMonitor component. Enforce hard rules to exit or hedge positions N hours before the earliest resolution window to avoid dispute lockup risk.

### ARCH-003: Cross-Market Arbitrage Requires a Constraint Graph
- **Severity:** Major
- **Source:** The Architect, Product Visionary
- **Description:** Strategy D requires understanding the logical relationships between markets (Temporal Containment, Mutual Exclusivity, Logical Implication). 
- **Recommendation:** Build a `ConstraintGraph` shared service that maps these relationships and constantly scans the CLOB for mathematical violations of these bounds.

### RISK-006: Leg Risk in CLOB Arbitrage
- **Severity:** Major
- **Source:** The Naysayer
- **Description:** Cross-market arbitrage assumes atomic execution. Polymarket CLOB is not atomic. If leg A fills and leg B fails, you hold naked directional risk.
- **Recommendation:** Design for leg risk: always place the less-liquid leg first, and implement an automated unwind sequence if the second leg fails to fill within milliseconds.

---

## Novel / Overlooked Strategies Identified

### 1. Bayesian Consistency Arbitrage (Conditional vs Unconditional)
- **Source:** The Naysayer
- **Description:** Polymarket often lists conditional markets ("Will X happen IF Y?") alongside unconditional ones ("Will X happen?"). The implied probability must satisfy Bayes' theorem: P(Outcome) = P(Cond) * P(Outcome|Cond) + P(Not Cond) * P(Outcome|Not Cond). When these drift out of mathematical sync, it creates a pure probabilistic arbitrage superior to simple correlation.

### 2. Resolution Timing Arbitrage (Time-Value Extraction)
- **Source:** The Architect
- **Description:** Markets price in a specific end date, but events often resolve early (e.g., a candidate drops out). The gap between $0.95 and $1.00 represents time-value. If you can model the *actual* expected resolution time vs the *market's* expected resolution time, you can extract a timing premium.

### 3. Probability-Surface Market Making
- **Source:** Creative Strategist
- **Description:** Instead of making markets on individual contracts, model a single "latent probability surface" for an entire cluster of related events. Quote bids and asks across all of them simultaneously derived from that single source of truth, harvesting incoherence from retail traders while keeping your overall inventory delta-neutral.

---

## What Was Done Well

- Mandating mathematically definable +EV instead of relying on pure low-latency or "gut feel" trading.
- Choosing Go for its concurrency (goroutines) and strict typing, which is vastly superior to Node/Python for managing stateful, high-frequency CLOB data.
- Identifying Cross-Market Correlation as a highly defensible edge that relies on logic rather than flawed ML forecasting.

---

## Action Items / The Execution Plan

- [ ] **Phase 0 (The Spike):** Build a Go WebSocket client to connect to Polymarket. Verify if historical data can be downloaded. If not, start recording the live WebSocket stream to disk immediately so you have data to backtest on next week.
- [ ] **Phase 0.5 (The Smoke Test):** Build the CLOB state manager and a pessimistic backtester (including fees and slippage). Implement **"Completeness Arbitrage"** (checking if mutually exclusive outcomes sum to >1 or <1) just to prove the engine can detect an edge and simulate a trade accurately.
- [ ] **Phase 1 (The Real Edge):** Build the Constraint Graph. Implement **Bayesian Consistency Arbitrage** and **Cross-Market Correlation Arbitrage**, as these offer the highest mathematical probability of success without relying on raw speed.

---

*Generated by DSS Architectural Review Panel · 2026-04-12*
