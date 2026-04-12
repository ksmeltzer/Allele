# Session 001: Initial Architecture & Language Selection
Date: 2026-04-12
Topic: Polymarket Arbitrage Backtesting Engine

## Context
- **Goal:** Build AI tools and a backtesting framework to test statistical theories and arbitrage opportunities on Polymarket.
- **Focus:** Mathematical and statistical arbitrage, expected value (+EV), and historical data simulation.

## Constraints
- **Language:** Python is strictly forbidden (enforced via Global Strata Rule).

## Active Decisions
- Core programming language for the live event-driven engine: **Go**. Chosen for optimal LLM generation, true concurrency (goroutines), and high performance.
- Strategy Focus: **Bayesian Consistency Arbitrage** and **Cross-Market Correlation Arbitrage**. These are mathematically bounded strategies that do not rely on raw speed or ML forecasting.
- Excluded Strategies: **Time-Decay (Theta) / Selling Miracles**. Strict risk constraint against asymmetric downside risk (picking up pennies in front of a steamroller).
- Strategy Optimization: **Genetic "Strategy Arena"**. The live engine will include a parameterized tournament system using genetic algorithms. Strategies will be assigned "genes" and will compete against live market data in parallel via Go routines.
- Risk Management: **Hard Stop-Losses & Global $100 Cap**. The system must utilize hard stop-losses. Furthermore, the absolute maximum capital allocation for the entire system is strictly capped at $100 for the live testing phase.
- Architecture Validation: **Completeness Arbitrage** (checking if mutually exclusive outcomes sum to exactly 1.0) will be used as the Phase 0 smoke test to validate data ingestion and fee modeling.
- Execution Realism: The execution engine MUST include pessimistic modeling of queue position, Polymarket fees, slippage, and "leg risk" (partial fills on multi-leg arbitrage).
- Live Micro-Trading Pivot: **Historical data recording is officially abandoned.** The market does not support it out-of-the-box. We are treating the first $100 as educational tuition and testing directly on the live Polymarket WebSocket stream.
