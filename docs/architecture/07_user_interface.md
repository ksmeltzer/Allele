# UI & Dashboard Architecture

## 1. Overview
The Allele UI is the **Global Command Center** for the Tri-Plugin Evolutionary ATS. Because the core engine abstracts away the complexity of multiple exchanges and niche data feeds, the frontend serves as a unified terminal to monitor the Genetic Arena, manage loaded plugins, and observe the live execution pipeline.

**Technical Constraints:**
*   **Framework:** React + Vite + Tailwind (Dark Mode).
*   **Communication:** A single Gorilla WebSocket connection to the Go backend (`ws://localhost:8081/stream`). REST polling is strictly prohibited.
*   **State:** The UI is a pure "dumb" subscriber. All state, credential encryption, and allocation math live in the Go backend.

---

## 2. Core Dashboard Views

### A. The Plugin Manager
This view parses and displays the contents of `.allele/plugins/<name>/` via the backend WebSocket.
*   **Exchanges:** Lists loaded Platform adapters (e.g., `polymarket.wasm`, `etrade.wasm`). Displays live WebSocket connection health (Green/Red indicator) and latency to the exchange.
*   **Sensors:** Lists loaded Data adapters (e.g., `twitter_sentiment.wasm`, `anthropic_llm.wasm`). Displays data-flow rates (e.g., "12 messages/sec" or "Polling every 60s").
*   **Strategies:** Lists mathematically pure logic modules. The UI parses the Strategy's exported `Manifest` to display a Dependency Graph (e.g., *"Requires: etrade, fda_rss. Status: All Dependencies Met."*).

### B. The Secure Vault (Account Configuration)
Because Strategies are pure and stateless, users must configure accounts at the Go Engine level.
*   **Credential Manager:** Form fields to add API keys for Sensors (e.g., OpenAI, Anthropic) and Exchanges (e.g., Binance API, Polymarket Private Keys).
*   **Transmission:** Keys entered here are sent immediately over the secure WebSocket to the Go Engine, which encrypts them and stores them in the SQLite `.allele/trading.db`. The UI never caches them locally.
*   **Global Capital Limits:** Inputs to define the system-wide maximum risk (e.g., "Global Cap: $100").

### C. The Genetic Arena (Leaderboard)
Replaces standard portfolio views. This tracks the performance of all living "Organisms" (Strategy + Sensors + Exchange + Genetic Parameters).
*   **Columns:** `Organism ID`, `Strategy WASM`, `Target Exchange`, `Sensors Used`.
*   **Fitness Metrics:** `Sortino Ratio`, `Max Drawdown`, `Win Rate`, `Capital Velocity`.
*   **Dynamic Allocation:** A visual representation of the Go Engine's Capital Allocator shifting the global $100 budget in real-time from losing organisms to the fittest organisms.

### D. The Global Radar (Live Execution Feed)
A unified, scrolling terminal combining the disparate async events of the microkernel into a single narrative timeline.
*   `[SENSOR]` Anthropic API: "News detected: FDA approves Pfizer."
*   `[STRATEGY]` Biotech_Arb: "Evaluating Snapshot... Edge detected: 14%."
*   `[VAULT]` Go Engine: "Unlocking credentials... Attaching Account_3 to Signal."
*   `[EXCHANGE]` E*TRADE: "Order placed. Filled at $45.20."
