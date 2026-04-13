package health

import (
	"context"
	"errors"
	"testing"
)

type mockPlugin struct {
	id      string
	pingErr error
}

func (m *mockPlugin) ID() string { return m.id }
func (m *mockPlugin) Ping(ctx context.Context) error { return m.pingErr }

func TestMonitor_PingPluginsAndTripBreaker(t *testing.T) {
	monitor := NewMonitor()
	
	// Create a failing plugin
	plugin1 := &mockPlugin{id: "plugin-db", pingErr: errors.New("timeout")}
	monitor.RegisterPlugin(plugin1, []string{"strategy-a", "strategy-b"})
	
	ctx := context.Background()
	
	// Ping should trigger failures. CB max failures is 3.
	monitor.PingPlugins(ctx) // Failure 1
	monitor.PingPlugins(ctx) // Failure 2
	
	if monitor.IsStrategyPaused("strategy-a") {
		t.Errorf("Expected strategy-a to NOT be paused before breaker trips")
	}
	
	monitor.PingPlugins(ctx) // Failure 3 - Breaker trips here
	
	if !monitor.IsStrategyPaused("strategy-a") {
		t.Errorf("Expected strategy-a to be paused after breaker trips")
	}
	if !monitor.IsStrategyPaused("strategy-b") {
		t.Errorf("Expected strategy-b to be paused after breaker trips")
	}
}

func TestMonitor_HaltDependentStrategies(t *testing.T) {
	monitor := NewMonitor()
	
	// Create a healthy plugin
	plugin2 := &mockPlugin{id: "plugin-api", pingErr: nil}
	monitor.RegisterPlugin(plugin2, []string{"strategy-c"})
	
	ctx := context.Background()
	monitor.PingPlugins(ctx)
	
	if monitor.IsStrategyPaused("strategy-c") {
		t.Errorf("Expected strategy-c to NOT be paused initially")
	}
	
	// Manually halt
	monitor.HaltDependentStrategies("plugin-api")
	
	if !monitor.IsStrategyPaused("strategy-c") {
		t.Errorf("Expected strategy-c to be paused after manual halt")
	}
}
