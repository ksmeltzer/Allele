# UI & Dashboard Architecture

## 1. Overview
The Allele UI is the **Global Command Center** for the Tri-Plugin Evolutionary ATS. Because the core engine abstracts away the complexity of multiple exchanges and niche data feeds, the frontend serves as a unified terminal to monitor the Genetic Arena, manage loaded plugins, and observe the live execution pipeline.

**Technical Constraints:**
*   **Framework:** React + Vite + Tailwind (Dark Mode).
*   **Communication:** A single Gorilla WebSocket connection to the Go backend (`ws://localhost:8081/stream`). REST polling is strictly prohibited.
*   **State:** The UI is a pure "dumb" subscriber. All state, credential encryption, and allocation math live in the Go backend.

---

## 2. Core Dashboard Views

### A. The Plugin Manager (Manifest & Dependencies)
This view manages the contents of the `plugins/` directory via the `/api/plugins` REST endpoints (supplementing the live WebSocket).
*   **WASM Manifests:** The UI fetches the `abi.Manifest` for every loaded plugin, displaying its type (Exchange, Sensor, Strategy), version, and author.
*   **Dependency Resolution Graph:** The UI parses the `Dependencies` array in the manifest. It must visually warn the user or block strategy execution if a required exchange or sensor plugin is not currently installed (e.g., *"Requires: allele-exchange-polymarket >= v1.0.0. Status: Missing."*).
*   **Live Metrics:** Displays WebSocket connection health (Green/Red indicator) and data-flow rates using the `/ws` stream.

### B. Dynamic Configuration & Secure Vault
Because the engine no longer uses `.env` files for security reasons, all system and plugin configurations are managed dynamically via the UI and stored in the backend SQLite `plugin_config` table.
*   **Dynamic Forms:** The UI uses the `ConfigField` schema from the `/api/plugins` endpoint to automatically generate configuration forms for each plugin (e.g., text inputs for API keys, toggles for experimental modes).
*   **Secret Masking:** Fields marked as `Type: "secret"` must be visually masked (`********`) when fetched from the backend.
*   **Submission:** User inputs are posted to the `/api/plugins/config` endpoint. The Go backend encrypts these secrets and stores them in the SQLite vault.
*   **Global System Settings:** Inputs to define the system-wide maximum risk (e.g., "Global Cap: $100") and disable the default Asymmetric Risk Ban.

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
