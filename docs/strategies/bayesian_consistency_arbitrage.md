# Bayesian Consistency Arbitrage

## Overview
Bayesian Consistency Arbitrage is a quantitative trading strategy that exploits statistical discrepancies between **conditional** and **unconditional** prediction markets.

Unlike simple correlation arbitrage or completeness arbitrage (which look at disjoint mutually exclusive events), Bayesian Consistency Arbitrage operates by verifying if the market's implied probabilities conform to the mathematical rules of probability theory, specifically Bayes' Theorem.

## The Mathematics
A prediction market often lists both unconditional questions and conditional questions:
* **Unconditional Market (A):** Will the Fed raise rates in 2024? (Let $P(A)$ be the implied probability of this happening).
* **Unconditional Market (B):** Will Bitcoin reach $100k in 2024? (Let $P(B)$).
* **Conditional Market (B|A):** *If* the Fed raises rates, will Bitcoin reach $100k? (Let $P(B|A)$).
* **Conditional Market (B|not A):** *If* the Fed does not raise rates, will Bitcoin reach $100k? (Let $P(B|\neg A)$).

According to the Law of Total Probability:
$$ P(B) = P(A) \cdot P(B|A) + P(\neg A) \cdot P(B|\neg A) $$

### The Arbitrage Opportunity
Because these markets trade independently, their prices are set by human supply and demand rather than a centralized math engine. This means that $P(B)$ can drift away from the mathematically correct value dictated by the other three markets.

If the market price of $P(B)$ is **greater** than the calculated expected value of $P(B)$ using the conditional markets, the unconditional market is *overpriced* relative to the conditional markets.

A trader can:
1. Sell (Short) the unconditional market $B$.
2. Buy the two conditional markets $B|A$ and $B|\neg A$ weighted by the probability of $A$.
3. Buy the unconditional market $A$ as a hedge.

If executed correctly, the trader locks in a profit regardless of whether $A$ happens or $B$ happens, because they have bought the "cheaper" synthetic version of $B$ and sold the "expensive" real version of $B$.

## The "Edge" (Why it makes money)
The edge here is superior statistical capability. Most retail traders and even some institutional algorithms trade markets in isolation. They might see news that makes them bearish on Bitcoin and sell $B$, without realizing they also need to update their positions in the conditional markets.

This creates a probabilistic arbitrage that is mathematically bound. It does not rely on forecasting *if* the Fed will raise rates. It only relies on the fact that the prices of these distinct markets *must* mathematically align.

## Real-World Complications (The Risks)

1. **Complex Hedging**: Constructing the exact synthetic hedge requires dynamic weighting. Since $P(A)$ constantly changes, the hedge ratio must be constantly rebalanced.
2. **Liquidity Constraints**: Conditional markets often have significantly lower liquidity (thinner order books) than unconditional markets. Slippage on the conditional legs can easily eat the arbitrage profit.
3. **Resolution Mechanics**: This strategy assumes all markets resolve using the exact same criteria. If the wording on market $A$ is slightly different than the condition required for $B|A$, the arbitrage breaks and the trader holds massive directional risk.
4. **Capital Lockup**: The trader must lock up capital in 4 separate markets simultaneously until resolution.

## Summary
Bayesian Consistency Arbitrage is a highly sophisticated, mathematically pure strategy that exploits the inefficiencies of human probability estimation across linked questions. It requires a robust execution engine capable of calculating dynamic hedge ratios and simultaneously managing liquidity across multiple order books.