# Genetic Algorithm & Strategy Arena

The Allele Engine utilizes a Genetic Algorithm (GA) to discover, optimize, and allocate capital to trading strategies. In algorithmic trading, the danger of an optimizer is "curve-fitting" (memorizing the past but failing in the future) and lacking diversification. 

This document outlines the architecture for the "Strategy Arena", ensuring fairness, robust out-of-sample performance, and an ensemble of **uncorrelated alphas** (strategies that make money at different times and in different ways).

## 1. The Anatomy of an "Organism" (The DNA)

In biology, an organism is an individual life form. In the Allele Engine, an organism (referred to technically as a **Phenotype** or **Strategy Instance**) is defined by a unique combination of three elements. Together, these form the organism's **DNA Fingerprint**:

1. **The Base Logic (The "Species")**: The underlying WASM binary file (e.g., `cross_market_arb.wasm`). This dictates the core rules of engagement and the mathematical logic.
2. **The Environment (The "Habitat")**: The specific markets or asset classes the organism is permitted to trade (e.g., US Political Markets, Crypto Price Markets, Pop Culture Markets).
3. **The Parameters (The "Genes")**: The array of tunable floating-point numbers the logic uses to make decisions. Examples include:
   * `min_profit_margin`: (e.g., 0.03 or 3%)
   * `max_hold_time_seconds`: (e.g., 3600)
   * `correlation_threshold`: (e.g., 0.95)

**The DNA Fingerprint** is generated as a cryptographic hash of `[WASM_ID + Market_Filter + Genes]`. Two organisms with the exact same DNA will theoretically make the exact same trades.

## 2. Ensemble Diversity (Uncorrelated Alphas)

If we spawn 1,000 organisms and simply take the Top 10 by P&L (Profit and Loss), we will likely select 10 nearly identical organisms that all traded the exact same market at the exact same time, just with slightly different margins. This is catastrophic for risk management. If that specific market type collapses, all 10 organisms die simultaneously.

To solve this, the GA enforces **Ensemble Diversity**.
*   **The Return Stream**: Every organism produces a "Return Stream" (an equity curve showing its P&L over time).
*   **Correlation Matrix**: The Arena calculates the Pearson correlation coefficient between the return streams of all top organisms.
*   **Correlation Penalty**: If Organism B is 95% correlated with Organism A (meaning they make and lose money at the exact same times), the GA heavily penalizes Organism B. 

The goal is to build a portfolio where Organism A makes money during high-volatility spikes, Organism B makes money during quiet weekends, and Organism C makes money purely on long-term resolution lockups. They are **uncorrelated**.

## 3. Fairness and The "Grace Period"

A strategy must be given a "fair shake." A brilliant long-term strategy might only make one trade a week, while a high-frequency strategy makes 100 trades a day. 

*   **Action-Based Burn-In**: An organism cannot be "killed" (culled) by the GA until it has reached a minimum threshold of *actionable events* (e.g., it must have evaluated at least 50 trade setups where its logic was triggered).
*   **Statistical Significance (p-value)**: The Fitness Function heavily weights statistical confidence. A strategy with a 100% win rate on 2 trades has a lower fitness score than a strategy with a 65% win rate on 200 trades.

## 4. The Fitness Function

The fitness function determines the "score" of an organism. It is mathematically designed to punish risk and reward consistency.

Instead of raw P&L, Allele uses an adjusted **Sortino Ratio** combined with capital efficiency:
1.  **Reward**: Annualized Expected Value (+EV).
2.  **Reward**: Capital Velocity (how quickly it frees up the $100 budget to be used again).
3.  **Penalty**: Maximum Drawdown (the largest peak-to-trough drop in its bankroll).
4.  **Penalty**: Downside Deviation (volatility that results in losses).
5.  **Penalty**: Correlation to the existing ensemble.

## 5. Lifecycle of an Organism

1. **Initialization**: The Arena spawns a population of 100 random organisms (random genes).
2. **Paper Trading (In-Sample)**: They are tested against historical JSONL ticks.
3. **Selection**: The top 20 uncorrelated organisms are selected.
4. **Crossover & Mutation**: Their "genes" are mathematically mixed to create 80 new offspring. A small random mutation (e.g., changing a 0.03 margin to 0.031) is applied to maintain genetic diversity.
5. **Walk-Forward Validation (Out-of-Sample)**: The survivors are tested on an *unseen* segment of data to prove they didn't just memorize the past.
6. **Live Allocation**: The best organisms are given a percentage of the live $100 bankroll. If they underperform in live trading, their allocation is slowly dialed down to $0 (hibernation), returning the capital to better-performing organisms.