package strategy

import (
	"log"

	"allele/internal/core"
)

type CompletenessArbitrage struct {
	marketOutcomes map[string][]string
	takerFee       float64
	minProfit      float64
}

func NewCompletenessArbitrage(takerFee float64, minProfit float64) *CompletenessArbitrage {
	return &CompletenessArbitrage{
		marketOutcomes: make(map[string][]string),
		takerFee:       takerFee,
		minProfit:      minProfit,
	}
}

func (c *CompletenessArbitrage) ID() string {
	return "completeness"
}

// RegisterMarket lets the strategy know which assets are mutually exclusive.
func (c *CompletenessArbitrage) RegisterMarket(conditionID string, assetIDs []string) {
	c.marketOutcomes[conditionID] = assetIDs
}

func (c *CompletenessArbitrage) Evaluate(state *core.MarketState) []core.Action {
	var actions []core.Action

	for conditionID, assetIDs := range c.marketOutcomes {
		totalCost := 0.0
		canTrade := true

		for _, assetID := range assetIDs {
			price, ok := state.AssetPrices[assetID]
			if !ok || price <= 0 {
				canTrade = false
				break
			}
			totalCost += price
		}

		if !canTrade {
			continue
		}

		totalCostWithFee := totalCost * (1.0 + c.takerFee)

		if 1.0-totalCostWithFee >= c.minProfit {
			log.Printf("[ARBITRAGE] Condition: %s | Cost (inc fee): %f | Profit/Share: %f", conditionID, totalCostWithFee, 1.0-totalCostWithFee)

			for _, assetID := range assetIDs {
				actions = append(actions, core.Action{
					MarketID:   "polymarket",
					AssetID:    assetID,
					Side:       core.BUY,
					Price:      state.AssetPrices[assetID], // We'd want to buy at best ask, which is the price we tracked
					Size:       1000000,                    // 1 Share
					StrategyID: c.ID(),
				})
			}
		}
	}

	return actions
}

func (c *CompletenessArbitrage) GetDNA() map[string]float64 {
	return map[string]float64{
		"takerFee":  c.takerFee,
		"minProfit": c.minProfit,
	}
}
