package engine

import (
	"context"
	"log"
	"sync"
	"time"

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
	EventBus    *core.EventBus
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
		EventBus:  core.NewEventBus(),
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
	if k.arena != nil {
		k.arena.AddOrganism(s.ID(), "WASM", "all", s.GetDNA())
	}
}

func (k *Kernel) TickChan() chan core.NormalizedTick {
	return k.tickChan
}

func (k *Kernel) Start(ctx context.Context) {
	if k.arena != nil {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Run the evolutionary selection and mutation process
					k.arena.Evolve(5, 20)
				}
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case tick := <-k.tickChan:
			k.MarketState.AssetPrices[tick.AssetID] = tick.Price

			if k.EventBus != nil {
				k.EventBus.Publish(core.Event{
					Type:    "tick",
					Payload: tick,
				})
			}

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

					if len(actions) > 0 && k.EventBus != nil {
						k.EventBus.Publish(core.Event{
							Type: "strategy_eval",
							Payload: map[string]interface{}{
								"strategy_id": strategyID,
								"actions":     actions,
							},
						})
					}

					if k.arena != nil && len(actions) > 0 {
						mu.Lock()
						// Mocking a small positive return per action so it has fitness to select on
						_ = k.arena.RecordAction(strategyID, float64(len(actions))*0.001)
						mu.Unlock()
					}

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
