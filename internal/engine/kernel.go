package engine

import (
	"context"
	"log"
	"sync"

	"allele/internal/arena"
	"allele/internal/core"
	"allele/internal/health"
)

type Kernel struct {
	exchanges   map[string]core.IExchange
	wallets     map[string]core.IWallet
	strategies  map[string]core.IStrategy
	MarketState *core.MarketState
	tickChan    chan core.NormalizedTick
	monitor     *health.Monitor
	arena       *arena.Arena
	sidelined   map[string]bool
}

func NewKernel() *Kernel {
	return &Kernel{
		exchanges:  make(map[string]core.IExchange),
		wallets:    make(map[string]core.IWallet),
		strategies: make(map[string]core.IStrategy),
		MarketState: &core.MarketState{
			AssetPrices: make(map[string]float64),
		},
		tickChan:  make(chan core.NormalizedTick, 1000),
		sidelined: make(map[string]bool),
	}
}

func (k *Kernel) SetMonitor(m *health.Monitor) {
	k.monitor = m
}

func (k *Kernel) SetArena(a *arena.Arena) {
	k.arena = a
}

func (k *Kernel) RegisterExchange(e core.IExchange) {
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

			var wg sync.WaitGroup
			mu := &sync.Mutex{}

			for _, strategy := range k.strategies {
				strategyID := strategy.ID()
				wg.Add(1)

				go func(strategy core.IStrategy, strategyID string) {
					defer wg.Done()

					if k.monitor != nil {
						if k.monitor.IsStrategyPaused(strategyID) {
							if k.arena != nil && !k.sidelined[strategyID] {
								mu.Lock()
								if err := k.arena.SidelineOrganism(strategyID); err != nil {
									log.Printf("Failed to sideline organism %s: %v", strategyID, err)
								} else {
									k.sidelined[strategyID] = true
								}
								mu.Unlock()
							}
							return
						} else if k.sidelined[strategyID] {
							if k.arena != nil {
								mu.Lock()
								if err := k.arena.ReactivateOrganism(strategyID); err != nil {
									log.Printf("Failed to reactivate organism %s: %v", strategyID, err)
								} else {
									k.sidelined[strategyID] = false
								}
								mu.Unlock()
							} else {
								k.sidelined[strategyID] = false
							}
						}
					}

					actions := strategy.Evaluate(k.MarketState)
					for _, action := range actions {
						exchangeID := action.MarketID
						if exchangeID == "" {
							exchangeID = "polymarket"
						}

						exchange, ok := k.exchanges[exchangeID]
						if !ok {
							log.Printf("Exchange not found: %s", exchangeID)
							continue
						}

						wallet, ok := k.wallets["polygon"]
						if !ok {
							log.Printf("Wallet not found: polygon")
							continue
						}

						if err := exchange.SubmitOrder(ctx, wallet, action); err != nil {
							log.Printf("Failed to submit order: %v", err)
						}
					}
				}(strategy, strategyID)
			}
			wg.Wait()
		}
	}
}
