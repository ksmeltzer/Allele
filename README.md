<div align="center">
  <img src="docs/assets/logo.svg" alt="Allele Logo" width="400">

  **A Universal Genetic Algorithmic Trading System (ATS).**
  
  <p>
    Built on a programmable WebAssembly microkernel, designed for cross-market micro-trading, Bayesian arbitrage, and genetic algorithm-driven strategy arenas.
  </p>

  [![Language: Go](https://img.shields.io/badge/Language-Go-00ADD8?style=flat-square&logo=go)](https://golang.org/)
  [![Runtime: WASM](https://img.shields.io/badge/Runtime-WebAssembly-654FF0?style=flat-square&logo=webassembly)](https://webassembly.org/)
  [![Markets: Universal](https://img.shields.io/badge/Targets-Polymarket_%7C_E*TRADE_%7C_Binance-2081E2?style=flat-square)](https://polymarket.com/)
  [![DB: SQLite](https://img.shields.io/badge/DB-Serverless_SQLite-003B57?style=flat-square&logo=sqlite)](https://modernc.org/sqlite)
</div>

---

## 🧬 Overview

**Allele** is a universal execution management system and programmable genetic trading platform built to operate across diverse financial and prediction markets. 

It uses a pure Go, zero-CGO backend executing hot-swappable WebAssembly (`.wasm`) logic modules, combined with a Genetic Algorithm (GA) "Arena" to continually evaluate, rank, and deploy "Organisms" (strategies + parameters + environments) based on their mathematical edge.

## 🏗️ Architecture Summary

Allele implements a **Hexagonal (Ports & Adapters) Microkernel Architecture**:

*   **The Firehose (Ingestion):** High-throughput raw market ticks stream directly to flat JSONL files in `.allele/historical/`.
*   **Live Micro-Trading:** Operates strictly on live WebSocket streams. REST polling for data is strictly prohibited.
*   **The Tri-Plugin Microkernel:** The engine dynamically loads three types of `.wasm` files from `.allele/plugins/<name>/` at runtime using `wazero` (a pure-Go WASM runtime):
    *   `exchanges/`: Adapters that translate proprietary market APIs (Polymarket, E*TRADE) into normalized `MarketStates`.
    *   `sensors/`: Adapters that fetch and normalize 3rd-party data (Twitter sentiment, Google Patents, LLMs).
    *   `strategies/`: Mathematically pure execution functions that subscribe to Exchanges and Sensors.
*   **The Secure Vault (Accounts):** The Go backend acts as a highly secure, centralized Vault holding all Exchange API keys and wallets. Strategies never see account credentials; the Go engine dynamically attaches credentials to the `Exchange` adapter strictly at the moment of execution.
*   **Serverless Storage:** Ledger, trades, and P&L aggregations use a CGO-free, serverless SQLite database (`modernc.org/sqlite`) stored locally at `.allele/trading.db`.
*   **DevOps & Monitoring:** Exclusively deployed via `podman-compose`. An independent OS-level Watchdog (`cmd/watchdog`) monitors the core engine via a TCP IPC socket and automatically recovers crashed, zombie, or detached containers.
*   **Dashboard:** A React + Vite + Tailwind Single Page Application (SPA), connected via a Gorilla WebSocket for real-time visualization of the orderbook and strategy radar.

---

## 🏛️ Strict Rules & Constraints

The Allele architecture enforces several immutable constraints to ensure stability and mitigate systemic risk:

### 1. No Black Boxes
Every strategy compiled to WASM **must** have an accompanying theoretical proof document in `docs/strategies/` explaining its exact mathematical edge. If a strategy's edge cannot be proven mathematically independently of how it is coded, it is forbidden from running.

### 2. Tri-WASM Architecture (Exchanges, Sensors & Strategies)
To support LLMs, proprietary stock markets, and AI without compromising the security or microsecond performance of the trading execution loop, the engine implements a 3-part WASM plugin architecture:
*   **Exchanges:** Granted WASI network access to connect to specific trading platforms.
*   **Sensors:** Granted restricted WASI network access to fetch 3rd-party data (e.g., Anthropic API). They *cannot* execute trades.
*   **Strategies:** Mathematically pure, zero-network execution functions. The Go engine acts as a Broker, piping data between the three.

### 3. Configurable Global Capital Cap
For the current live testing phase, the absolute maximum capital allocation for the entire system is configured by default to a conservative **$100** cap. This is fully configurable by the user via the system settings..

### 4. Asymmetric Risk Ban
Strategies with unbounded or highly asymmetric downside risk (e.g., betting heavily against low-probability "miracle" outcomes for tiny "Theta decay" gains) are strictly prohibited. 

### 5. Global Ensemble Diversity (Genetic Arena)
The GA fitness function enforces "Ensemble Diversity" across all active exchanges simultaneously. The Capital Allocator will ruthlessly defund an organism trading Polymarket if an organism trading E*TRADE demonstrates a higher risk-adjusted return (Sortino Ratio). The goal is to build a global basket of uncorrelated alphas.

---

## 🚀 Strategy Development Lifecycle

Strategies are pure mathematical functions that evaluate a normalized `MarketState` and output an array of `Signal` structs (`[]Signal`). They **never** place orders directly. 

1. Write the theory in `docs/strategies/`.
2. Implement the strategy logic against the `Strategy` interface.
3. Compile to WASM using `GOOS=wasip1 GOARCH=wasm`.
4. Drop the `.wasm` file into `.allele/strategies/`.
5. The Genetic Arena automatically benchmarks the new strategy's return stream.

*(Example: Completeness Arbitrage serves as the architectural smoke test in `strategies/src/completeness/`)*

---

## 💻 Operations & CLI

Allele abstracts away all underlying infrastructure (Podman, SQLite, WebSockets) through a unified CLI. 

All state lives entirely within the hidden `.allele/` directory.

```bash
# Initialize the Allele environment, start Podman containers, and launch Watchdog
allele start

# Deploy a newly compiled strategy
allele deploy path/to/strategy.wasm

# Inspect real-time performance and view the UI
allele ui
```

*(Note: Raw `podman` or `docker` commands should be avoided. Always use the `allele` wrapper.)*

---

> **Disclaimer:** This system is for research and educational purposes regarding genetic algorithms and prediction market arbitrage. Usage of real funds carries inherent risk of loss due to market volatility, oracle disputes, smart contract vulnerabilities, and unexpected protocol changes.
