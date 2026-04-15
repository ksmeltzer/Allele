package strategy

import (
	"log"
	"strconv"
	"sync"

	"allele/internal/core"
)

type CompletenessArbitrage struct {
	marketOutcomes map[string][]string
	takerFee       float64
	minProfit      float64
	eventBus       *core.EventBus
	mu             sync.RWMutex
}

func NewCompletenessArbitrage(takerFee float64, minProfit float64, eventBus *core.EventBus) *CompletenessArbitrage {
	c := &CompletenessArbitrage{
		marketOutcomes: make(map[string][]string),
		takerFee:       takerFee,
		minProfit:      minProfit,
		eventBus:       eventBus,
	}

	if eventBus != nil {
		go c.listenForConfig()
	}

	return c
}

func (c *CompletenessArbitrage) listenForConfig() {
	ch := c.eventBus.Subscribe(core.ConfigUpdatedEvent)
	for event := range ch {
		payload, ok := event.Payload.(map[string]string)
		// Assume the UI configuration specifies the plugin name for this strategy
		if !ok || payload["plugin_name"] != "allele-strategy-completeness" {
			continue
		}

		log.Println("CompletenessArbitrage: Config updated, reloading parameters...")

		c.mu.Lock()
		if tfStr, ok := payload["TAKER_FEE"]; ok {
			if tf, err := strconv.ParseFloat(tfStr, 64); err == nil {
				c.takerFee = tf
			}
		}
		if mpStr, ok := payload["MIN_PROFIT"]; ok {
			if mp, err := strconv.ParseFloat(mpStr, 64); err == nil {
				c.minProfit = mp
			}
		}
		c.mu.Unlock()

		c.eventBus.Publish(core.Event{
			Type: core.SystemAlertEvent,
			Payload: map[string]interface{}{
				"source":  "allele-strategy-completeness",
				"level":   "info",
				"message": "Completeness strategy parameters updated successfully.",
			},
		})
	}
}

func (c *CompletenessArbitrage) ID() string {
	return "completeness"
}

// RegisterMarket lets the strategy know which assets are mutually exclusive.
func (c *CompletenessArbitrage) RegisterMarket(conditionID string, assetIDs []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.marketOutcomes[conditionID] = assetIDs
}

func (c *CompletenessArbitrage) Evaluate(state *core.MarketState) []core.Action {
	var actions []core.Action

	c.mu.RLock()
	defer c.mu.RUnlock()

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
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]float64{
		"takerFee":  c.takerFee,
		"minProfit": c.minProfit,
	}
}
