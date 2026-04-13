# Strategy Library

This directory serves as an educational library of quantitative trading strategies. 

Each document outlines the mathematical and conceptual foundation of a specific trading strategy, completely independent of how it might be implemented in code. This allows for a deeper understanding of the "edge" (why the strategy makes money) and the theoretical risks involved.

## 🚨 Architectural Rule: No Black Boxes

The Allele Engine strictly enforces a **"No Black Boxes"** policy.
Every WASM trading plugin deployed to the system **MUST** have an accompanying markdown document in this library. If a strategy's mathematical edge, expected value (+EV) logic, and associated risks cannot be clearly explained in plain English and math independent of the code, the plugin is considered invalid and will not be permitted to trade live capital.

## Available Strategies

1. **[Completeness Arbitrage](completeness_arbitrage.md)**: The foundational "smoke test" of prediction markets—exploiting markets where mutually exclusive outcomes do not sum to exactly 100%. *(Example Plugin Available in `strategies/src/completeness`)*
2. **[Bayesian Consistency Arbitrage](bayesian_consistency_arbitrage.md)**: Exploiting discrepancies between conditional and unconditional probabilities using Bayes' Theorem.
3. **[Cross-Market Correlation Arbitrage](cross_market_correlation_arbitrage.md)**: Hedging and exploiting highly correlated but fundamentally distinct markets.
4. **[Resolution Timing Arbitrage](resolution_timing_arbitrage.md)**: Capitalizing on the time-value of money and capital lockups during market resolution periods.
