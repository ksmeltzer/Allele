package engine

import (
	"context"
	"testing"
	"time"

	"allele/internal/arena"
	"allele/internal/core"
	"allele/internal/health"
)

type mockStrategy struct {
	id            string
	evaluateCount int
}

func (m *mockStrategy) ID() string { return m.id }
func (m *mockStrategy) Evaluate(state *core.MarketState) []core.Action {
	m.evaluateCount++
	return nil
}
func (m *mockStrategy) GetDNA() map[string]float64 { return nil }

type dummyPlugin struct {
	id string
}

func (p *dummyPlugin) ID() string                     { return p.id }
func (p *dummyPlugin) Ping(ctx context.Context) error { return nil }

func TestKernel_HealthArenaIntegration(t *testing.T) {
	kernel := NewKernel()

	mon := health.NewMonitor()
	ar := arena.NewArena()
	kernel.SetMonitor(mon)
	kernel.SetArena(ar)

	strategyID := "test-strategy-1"
	ar.AddOrganism(strategyID, "test_species", "test_filter", nil)

	strategy := &mockStrategy{id: strategyID}
	kernel.RegisterStrategy(strategy)

	plugin := &dummyPlugin{id: "dummy-plugin"}
	mon.RegisterPlugin(plugin, []string{strategyID})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go kernel.Start(ctx)

	time.Sleep(10 * time.Millisecond)

	kernel.TickChan() <- core.NormalizedTick{AssetID: "ETH", Price: 2000.0}

	time.Sleep(50 * time.Millisecond)

	if strategy.evaluateCount != 1 {
		t.Errorf("Expected evaluateCount to be 1, got %d", strategy.evaluateCount)
	}

	org, err := ar.GetOrganism(strategyID)
	if err != nil {
		t.Fatalf("Failed to get organism: %v", err)
	}
	if org.IsSidelined {
		t.Errorf("Organism should not be sidelined initially")
	}

	mon.HaltDependentStrategies(plugin.ID())

	kernel.TickChan() <- core.NormalizedTick{AssetID: "ETH", Price: 2005.0}

	time.Sleep(50 * time.Millisecond)

	if strategy.evaluateCount != 1 {
		t.Errorf("Expected evaluateCount to remain 1, got %d", strategy.evaluateCount)
	}

	if !org.IsSidelined {
		t.Errorf("Organism should be sidelined after being paused")
	}

	// Since monitor doesn't have an unpause method, we swap out the monitor
	// for a fresh one to simulate the strategy being unpaused
	mon2 := health.NewMonitor()
	kernel.SetMonitor(mon2)

	// Another tick should reactivate it
	kernel.TickChan() <- core.NormalizedTick{AssetID: "ETH", Price: 2010.0}

	time.Sleep(50 * time.Millisecond)

	if strategy.evaluateCount != 2 {
		t.Errorf("Expected evaluateCount to increment to 2, got %d", strategy.evaluateCount)
	}

	if org.IsSidelined {
		t.Errorf("Organism should be reactivated after being unpaused")
	}
}
