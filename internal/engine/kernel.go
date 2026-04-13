package engine

import (
	"context"
	"log"

	"allele/internal/core"
)

type Kernel struct {
	exchanges   map[string]core.IExchange
	wallets     map[string]core.IWallet
	strategies  map[string]core.IStrategy
	MarketState *core.MarketState
	tickChan    chan core.NormalizedTick
}

func NewKernel() *Kernel {
	return &Kernel{
		exchanges:  make(map[string]core.IExchange),
		wallets:    make(map[string]core.IWallet),
		strategies: make(map[string]core.IStrategy),
		MarketState: &core.MarketState{
			AssetPrices: make(map[string]float64),
		},
		tickChan: make(chan core.NormalizedTick, 1000),
	}
}

func (k *Kernel) RegisterExchange(e core.IExchange) {
	// For simplicity, using "polymarket" as ID for now.
	// A better way is if IExchange has an ID() method, but core.IExchange doesn't.
	k.exchanges["polymarket"] = e
}

func (k *Kernel) RegisterWallet(w core.IWallet) {
	k.wallets["polygon"] = w
}

func (k *Kernel) RegisterStrategy(s core.IStrategy) {
	k.strategies[s.ID()] = s
}

func (k *Kernel) TickChan() chan core.NormalizedTick {
	return k.tickChan
}

func (k *Kernel) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case tick := <-k.tickChan:
			k.MarketState.AssetPrices[tick.AssetID] = tick.Price

			for _, strategy := range k.strategies {
				actions := strategy.Evaluate(k.MarketState)
				for _, action := range actions {
					exchangeID := action.MarketID
					if exchangeID == "" {
						exchangeID = "polymarket" // fallback
					}

					exchange, ok := k.exchanges[exchangeID]
					if !ok {
						log.Printf("Exchange not found: %s", exchangeID)
						continue
					}

					// We use "polygon" wallet by default for now
					wallet, ok := k.wallets["polygon"]
					if !ok {
						log.Printf("Wallet not found: polygon")
						continue
					}

					err := exchange.SubmitOrder(ctx, wallet, action)
					if err != nil {
						log.Printf("Failed to submit order: %v", err)
					}
				}
			}
		}
	}
}
