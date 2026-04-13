# Requirements

## 1. Business Requirements
* **Goal**: Build an AI-driven backtesting and live-trading framework to test statistical theories and execute arbitrage opportunities on prediction markets (specifically Polymarket).
* **Focus**: Mathematical and statistical arbitrage, expected value (+EV), and historical data simulation.
* **Risk Tolerance**: Strict adherence to capital preservation. The absolute maximum capital allocation for the entire system is configured by default to $100 but can be configured by the user for the initial live testing phase. Strategies with unbounded or highly asymmetric downside risk (e.g., selling low-probability "miracle" outcomes for tiny gains) are banned by default, but this can be disabled in the user configuration for experimental strategies.
* **Operational Mode**: The system must operate entirely within the hidden `.allele` directory for all local data, SQLite DBs, and watchdog binaries.

## 2. Functional Requirements
* **Strategy Engine**: Must support evaluating normalized `MarketState` data and outputting `Action` structs (pure functions). The strategy evaluation must be pluggable using a WebAssembly (WASM) runtime (`wazero`) capable of hot-loading `.wasm` binaries without recompiling the core application.
* **Execution Simulator**: Must include pessimistic modeling of queue position, exchange fees (e.g., 2% taker fee), slippage, and "leg risk" (partial fills on multi-leg arbitrage).
* **Live Micro-Trading**: Must operate directly on live WebSocket streams (REST polling is strictly prohibited).
* **Arbitrage Types**: 
  * Phase 0: Completeness Arbitrage (Smoke Test)
  * Phase 1: Cross-Market Correlation Arbitrage
  * Phase 2: Bayesian Consistency Arbitrage
* **Order Management**: Must use EIP-712 signed limit orders and HMAC-SHA256 REST auth. Limit orders use a 2-second expiration to simulate Fill-or-Kill (FOK) and mitigate leg risk. Gasless position merging uses the Relayer API.
* **Risk Management**: Stop-losses are mandatory for all algorithmic trading strategies. The maximum possible loss must be strictly capped to the allocated capital ("the pot") for that specific trade.

## 3. Non-Functional Requirements
* **Performance**: The primary programming language for the backtesting engine and data ingestion pipeline is Go (Golang). Python is strictly forbidden.
* **Concurrency**: Must leverage Go's true concurrency (goroutines) for infinite parallel backtesting and genetic tournament optimization without network calls.
* **Architecture**: Must use a Hexagonal (Ports and Adapters) Microkernel architecture. Core interfaces (`IExchange`, `IWallet`, `IStrategy`, `IEngine`) are completely decoupled.
* **Storage**: High-throughput raw market ticks stream directly to flat JSONL files in `.allele/historical/` (The Firehose). The genetic strategy ledger, trades, and P&L aggregations use a serverless SQLite database (`modernc.org/sqlite` without CGO requirements) located at `.allele/trading.db`.
* **UI**: The frontend is a dark-mode React + Vite + Tailwind Single Page Application (SPA) served via an Nginx container. It connects to the Go backend exclusively via a Gorilla WebSocket on port `8081`.
* **DevOps**: Podman is exclusively used for containerization (`podman-compose`). A decoupled native OS Watchdog (`cmd/watchdog`) monitors the trading engine via a TCP IPC socket (port 9999).

## 4. Architectural Constraints
* **No Official Client**: The Polymarket CLOB API has no official Go client. A native Go client must be built using the Rust client as the reference specification.
* **Oracle Risk**: The engine must detect and handle UMA Oracle resolution disputes and enforce mandatory exit windows before resolution.
* **Collateral Risk**: The engine must detect and handle both standard and negRisk markets, which have different collateral mechanics.
* **Execution**: Strategies must never place orders directly — signal generation and execution are strictly separated. Each `Signal` must carry ExpectedEV, Confidence, and Reason fields for auditability.