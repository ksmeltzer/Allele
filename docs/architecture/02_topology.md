# System Topology

## High-Level Container Topology

```mermaid
graph TD
    subgraph Host OS [Local Development Environment]
        Watchdog(Native Go Daemon: cmd/watchdog)
        Browser(Allele UI SPA)
        Filesystem[(~/.allele/)]
        SQLite[(~/.allele/trading.db)]
        JSONL[(~/.allele/historical/)]
    end

    subgraph Podman Compose [Container Network]
        Engine[Allele Trading Engine (Go)]
        Nginx[Allele UI Server]
    end

    subgraph External Networks [Polymarket APIs]
        CLOB_WS(CLOB WebSocket)
        Relayer_API(Relayer REST API)
        CLOB_REST(CLOB REST API)
        Polygon(Polygon Chain 137 RPC)
    end

    subgraph External Notifications
        Telegram(Telegram API)
    end

    %% Internal Connections
    Engine -- "Raw Ticks (I/O)" --> JSONL
    Engine -- "Strategy/Ledger (I/O)" --> SQLite
    Engine -- "TCP Heartbeats/Crashes (9999)" --> Watchdog
    Watchdog -- "Alerts" --> Telegram

    %% Frontend Connections
    Browser -- "HTTP :80" --> Nginx
    Nginx -- "Static Assets" --> Browser
    Browser -- "WS :8081 (Orderbooks, Radar)" --> Engine

    %% External Data Ingestion & Execution
    CLOB_WS -- "Live Ticks" --> Engine
    Engine -- "HMAC-SHA256 Auth / Orders" --> CLOB_REST
    Engine -- "EIP-712 Signatures" --> Polygon
    Engine -- "Gasless Position Merging" --> Relayer_API
```

## Internal Engine Process Flow

```mermaid
graph LR
    subgraph Ticks [Data Ingestion]
        WS[WebSocket Stream]
    end

    subgraph Core [Hexagonal Core]
        NormalizedTick(Raw Market Ticks)
        MarketState(Normalized State Book)
        Strategy[IStrategy.Evaluate]
        Signal[[]Signal]
    end

    subgraph Execution [Order Management]
        Risk[Risk Management & PnL]
        Orders(Order Execution Service)
        Action[[]Action]
    end

    WS --> NormalizedTick
    NormalizedTick --> MarketState
    MarketState --> Strategy
    Strategy --> Signal
    Signal --> Risk
    Risk --> Action
    Action --> Orders
```