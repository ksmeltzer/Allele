# Functional Architecture

## 1. Data Ingestion (The Firehose)
*   **CLOB WebSocket Client**: The system connects to the Polymarket CLOB API via a native Go client built in `internal/polymarket/`, using the Rust client as the reference specification. It streams raw market ticks directly to flat JSONL files in `.allele/historical/` (The Firehose).
*   **Normalization**: The raw ticks are converted to a `NormalizedTick` format before being passed to the core engine.

## 2. Core Engine (Microkernel)
*   **State Management**: The core engine maintains a normalized `MarketState` object representing the current order books, implied probabilities, and user portfolio balances.
*   **Strategy Evaluation**: The core engine evaluates all registered strategies (implementing `IStrategy`) against the `MarketState`. Each strategy is a pure function that returns a slice of `Signal` structs, carrying ExpectedEV, Confidence, and Reason fields for auditability.
*   **Execution Simulator**: For backtesting, the core engine simulates execution by applying pessimistic modeling of queue position, 2% taker fees, slippage, and "leg risk". It outputs `Action` structs.

## 3. Order Execution (The Trader)
*   **EIP-712 Order Signing**: The execution service handles EIP-712 order signing on Polygon chain 137. It uses a 2-second expiration to simulate Fill-or-Kill (FOK) and mitigate leg risk.
*   **HMAC-SHA256 Auth**: The execution service authenticates with the Polymarket CLOB REST API.
*   **Risk Management**: Stop-losses are mandatory. The maximum possible loss is strictly capped to the allocated capital ("the pot") for that specific trade. The global system maximum cap defaults to $100 but is configurable.
*   **Gasless Position Merging**: The execution service uses the Polymarket Relayer API to manage negRisk markets and merge positions.

## 4. UI & Dashboard (The Radar)
*   **SPA**: A dark-mode React + Vite + Tailwind Single Page Application (SPA) served via an Nginx container (`allele_ui`).
*   **Gorilla WebSocket**: Connects to the Go backend exclusively via a Gorilla WebSocket on port `8081` to display real-time orderbooks, radar updates, and execution feeds. No REST polling is permitted.

## 5. DevOps & Monitoring (The Watchdog)
*   **Native Daemon**: A decoupled native OS Watchdog (`cmd/watchdog`) monitors the trading engine via a TCP IPC socket (port 9999).
*   **Heartbeats & Crashes**: The trading engine streams `EventHeartbeat` and `EventCrash` to the Watchdog.
*   **Alerting**: The Watchdog handles Telegram crash alerts and issues `podman restart` if the socket drops or heartbeats fail.
*   **Containerization**: Podman is exclusively used for containerization (`podman-compose`). The `allele_engine` and `allele_ui` containers are managed by the `allele` CLI script.