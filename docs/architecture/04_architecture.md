# Architectural Design

## Overview
The Allele Trading Engine is a high-performance, purely concurrent, deterministic quantitative trading system designed explicitly for prediction markets (Polymarket). It is built entirely in Go (Golang) to support infinite parallel backtesting and genetic tournament optimization without network calls.

The system abandons traditional REST polling in favor of a 100% event-driven, WebSocket-first architecture. It operates entirely within a hidden `.allele` directory for all local data, SQLite DBs, and watchdog binaries.

## 1. Hexagonal (Ports and Adapters) Microkernel Architecture
The core system is structured around a Hexagonal Microkernel architecture. The business logic (the "Kernel") is completely decoupled from external dependencies.

*   **Core Interfaces**: `IExchange`, `IWallet`, `IStrategy`, `IEngine` define the strict contracts that all adapters must implement.
*   **Pure Functions**: Strategies (`IStrategy.Evaluate`) are pure mathematical functions. They take an immutable `MarketState` snapshot as input and return an array of `Signal` structs. They have zero side effects and make zero network calls.
*   **Separation of Concerns**: Strategies generate `Signal`s (intent). The Execution layer converts `Signal`s into `Action`s (execution). The Adapters translate `Action`s into API calls (e.g., EIP-712 signatures). This separation guarantees that a strategy cannot accidentally drain a wallet due to a bug in order formatting.

## 2. Hybrid Storage Model
The system employs a two-tier storage architecture optimized for different workloads:

*   **The Firehose (High-Throughput I/O)**: Raw market ticks streamed from the CLOB WebSocket are written directly to append-only, flat JSONL files in `.allele/historical/`. This avoids database write contention during extreme market volatility.
*   **The Ledger (Structured Querying)**: The genetic strategy ledger, trade history, P&L aggregations, and system configuration are stored in a serverless SQLite database (`modernc.org/sqlite` — chosen specifically to avoid CGO requirements, ensuring cross-platform portability of the Go binary). This database lives at `.allele/trading.db`.

## 3. Pessimistic Execution Simulator
Because historical backtesting on Polymarket is notoriously difficult due to lack of public APIs, the engine implements a highly pessimistic execution simulator designed to rigorously test strategies against realistic market conditions.

*   **Fee Modeling**: All expected values (+EV) are calculated *after* applying the maximum Polymarket taker fee (e.g., 2%).
*   **Queue Position**: The simulator does not assume mid-market fills. It models order book depth and assumes orders are placed at the back of the queue.
*   **Leg Risk Simulation**: For multi-leg strategies (like Cross-Market Correlation Arbitrage), the simulator explicitly models the probability of partial fills, where one leg executes but the other fails, leaving the system with naked directional risk.

## 4. UI & DevOps
*   **Event-Driven UI**: The frontend (React + Vite + Tailwind) is a pure subscriber. It connects to the Go backend via a single Gorilla WebSocket on port `8081`. It receives real-time order book updates, radar pings, and execution feeds. It never polls the backend via REST.
*   **IPC Watchdog**: To prevent silent container failures ("zombie containers"), the system uses a decoupled, native OS daemon (`cmd/watchdog`). The Go engine (`allele_engine` running in Podman) connects to the watchdog via a local TCP socket (`127.0.0.1:9999`). 
    *   The engine streams a heartbeat every 5 seconds (only if it has recently received WS ticks, proving the external connection is alive).
    *   If the engine panics, a deferred function catches it and pipes a CRASH event over the socket before dying.
    *   If the watchdog misses heartbeats or receives a crash, it alerts the user via Telegram and issues a `podman restart allele_engine` command.
## 5. Pluggable WebAssembly (WASM) Strategy Engine
To allow infinite flexibility without modifying the core Go codebase, the Allele engine utilizes a **WebAssembly (WASM) Plugin Architecture**.

*   **Dynamic Loading**: Strategies are not hardcoded into the Go binary. Instead, the engine dynamically loads `.wasm` files from the `.allele/strategies/` directory at runtime.
*   **Pure Go Runtime**: The engine embeds `wazero`, a zero-dependency, pure-Go WebAssembly runtime. This ensures that the entire system remains easily deployable across any OS/Arch without requiring CGO or external C compilers.
*   **Language Agnostic**: Because strategies compile to WASM, quantitative developers can write their alpha logic in Rust, Go, AssemblyScript, Python, or C++. As long as the resulting WASM binary implements the `Evaluate(statePtr) -> signalsPtr` memory contract, the Go host will execute it.
*   **Hot-Swapping**: The engine can watch the `.allele/strategies/` directory and automatically hot-swap strategy logic or parameters without restarting the WebSocket feeds or risking downtime.
*   **Sandboxing**: WASM provides a strictly sandboxed execution environment. A faulty strategy cannot read the host filesystem, access the network, or crash the core trading engine with a panic. It can only compute numbers and return Signals.
