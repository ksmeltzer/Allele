# Glossary of Terms

This glossary provides definitions for the key concepts, acronyms, and jargon used throughout the Allele project. It is categorized by industry to help bridge the gap between software engineering, quantitative finance, and evolutionary biology.

| Term | Industry | Definition |
| :--- | :--- | :--- |
| **Alpha** | Financial | The excess return of an investment relative to the return of a benchmark index. In algorithmic trading, "finding alpha" means discovering a strategy that consistently generates profit above what could be made passively (e.g., buying and holding the S&P 500). |
| **Arbitrage** | Financial | The simultaneous purchase and sale of an asset to profit from an imbalance in the price. It is theoretically a "risk-free" trade. |
| **CLOB** | Financial | Central Limit Order Book. A transparent system that matches customer bids and offers on an exchange. |
| **Chromosome** | Biology / GA | The entire set of "Genes" (parameters) that define an individual "Organism" (trading bot) in a Genetic Algorithm. |
| **Circuit Breaker** | Risk / Ops | A safety mechanism in the Go Engine. If a Sensor or Exchange plugin fails its heartbeat or returns errors, the circuit breaker "trips" and halts all dependent Strategies to prevent rogue trading on bad data. |
| **Crossover (Recombination)** | Biology / GA | A genetic operator used in the Genetic Algorithm to combine the genetic information of two parents to generate new offspring (e.g., taking half the parameters of the best-performing bot and half from the second-best to create a new one). |
| **Drawdown (Max Drawdown)** | Financial | The peak-to-trough decline of an investment's value during a specific period. It is a primary measure of downside risk. A 50% drawdown means the bot lost half its money before recovering. |
| **Edge** | Common Use | A competitive advantage. In trading, it is the mathematical or statistical reason *why* a strategy expects to make money over thousands of trades. |
| **Ensemble** | Common Use | A group of items viewed as a whole. In Allele, an "Ensemble of Alphas" refers to a diversified portfolio of different trading bots running simultaneously to reduce overall system risk. |
| **EV (Expected Value)** | Mathematics | The anticipated value for an investment at some point in the future, calculated by multiplying each possible outcome by the probability of its occurrence and summing all of those values. |
| **Exchange Plugin** | Technical / Tri-Plugin | A WASM adapter that translates proprietary exchange APIs (Polymarket, E*TRADE) into normalized `MarketStates` and executes standard `TradeSignals`. |
| **Fitness Function** | Biology / GA | The mathematical formula used to evaluate how "good" an organism is. In Allele, this is the scoring system (e.g., high returns, low risk, low correlation) that determines which bots survive and reproduce. |
| **Gene** | Biology / GA | A single tunable parameter within a strategy (e.g., `min_profit_margin=0.03` or `max_hold_time=3600`). |
| **Hexagonal Architecture** | Software Eng. | An architectural pattern (also known as Ports and Adapters) that isolates the core logic of an application from outside concerns (like databases, UIs, or external APIs). |
| **JSONL** | Technical | JSON Lines. A text format where each line is a valid JSON object. Used by Allele for high-throughput streaming of historical market data (The Firehose) because it is faster to append to than a traditional database. |
| **Limit Order** | Financial | An order placed with a brokerage to buy or sell a set number of shares at a specified price or better. It guarantees the price but does not guarantee execution (unlike a Market Order). |
| **Maker / Taker** | Financial | "Makers" provide liquidity to an order book by placing orders that sit on the book. "Takers" remove liquidity by placing orders that match immediately with existing orders. Takers usually pay higher fees. |
| **Organism (Phenotype)** | Biology / GA | A single, fully-realized trading bot instance. It is the combination of the WASM Strategy Logic (The Species), the Market Filters (The Habitat), and the Parameters (The Genes). |
| **Sensor Plugin** | Technical / Tri-Plugin | A WASM adapter granted restricted network access to fetch and normalize 3rd-party data (e.g., Twitter sentiment, Google Patents) into `DataPayloads` for strategies to consume. |
| **Sharpe / Sortino Ratio** | Financial | Mathematical formulas used to calculate risk-adjusted return. They penalize strategies that take massive risks to make small profits. The Sortino ratio is favored by Allele because it only penalizes *downside* volatility (losses), whereas Sharpe penalizes all volatility (including massive unexpected profits). |
| **Slippage** | Financial | The difference between the expected price of a trade and the price at which the trade is actually executed. Often occurs in thin/low-liquidity markets. |
| **Strategy Plugin** | Technical / Tri-Plugin | A mathematically pure, zero-network WASM execution function. It subscribes to Exchanges and Sensors and emits `TradeSignals`. |
| **Tri-Plugin Architecture** | Software Eng. | Allele's core microkernel design separating concerns into Exchanges (Execution), Sensors (Data), and Strategies (Logic). |
| **Vault** | Security | The highly secure, centralized Go Engine component that stores actual Account Credentials (API keys, OAuth tokens). It dynamically attaches these to standard `TradeSignals` at the moment of execution, keeping Strategies entirely unaware of keys. |
| **WASM (WebAssembly)** | Technical | A binary instruction format for a stack-based virtual machine. It allows code written in languages like Rust, C++, or Go to run at near-native speed in secure, sandboxed environments (like a web browser or the Allele backend engine). |