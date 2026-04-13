package health

import (
	"context"
	"sync"
	"time"
)

// Plugin represents a loadable extension that can be monitored for health.
type Plugin interface {
	ID() string
	Ping(ctx context.Context) error
}

// Monitor tracks the health of registered plugins and halts dependent strategies on failure.
type Monitor struct {
	mu               sync.RWMutex
	plugins          map[string]Plugin
	breakers         map[string]*CircuitBreaker
	pausedStrategies map[string]bool
	pluginStrategies map[string][]string // Maps plugin ID to a list of dependent strategy IDs
}

// NewMonitor initializes a new health monitor.
func NewMonitor() *Monitor {
	return &Monitor{
		plugins:          make(map[string]Plugin),
		breakers:         make(map[string]*CircuitBreaker),
		pausedStrategies: make(map[string]bool),
		pluginStrategies: make(map[string][]string),
	}
}

// RegisterPlugin adds a plugin to the monitor, along with the strategies that depend on it.
func (m *Monitor) RegisterPlugin(p Plugin, dependentStrategies []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins[p.ID()] = p
	// Trip circuit breaker after 3 failures
	m.breakers[p.ID()] = NewCircuitBreaker(3, time.Minute)
	m.pluginStrategies[p.ID()] = dependentStrategies
}

// PingPlugins loops through all registered plugins, pings them, and updates circuit breakers.
func (m *Monitor) PingPlugins(ctx context.Context) {
	// Snapshot the map to avoid holding the lock during blocking Ping calls.
	m.mu.RLock()
	pluginsCopy := make(map[string]Plugin, len(m.plugins))
	for id, p := range m.plugins {
		pluginsCopy[id] = p
	}
	m.mu.RUnlock()

	for id, p := range pluginsCopy {
		err := p.Ping(ctx)
		m.mu.Lock()
		cb := m.breakers[id]
		if err != nil {
			cb.RecordFailure()
			if cb.IsOpen() {
				m.haltDependentStrategiesLocked(id)
			}
		} else {
			cb.RecordSuccess()
		}
		m.mu.Unlock()
	}
}

// HaltDependentStrategies manually trips the circuit breaker for a plugin and halts its dependent strategies.
func (m *Monitor) HaltDependentStrategies(pluginID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cb, exists := m.breakers[pluginID]; exists {
		cb.Trip()
	}
	m.haltDependentStrategiesLocked(pluginID)
}

// haltDependentStrategiesLocked halts all strategies dependent on the specified plugin.
// Must be called with m.mu locked.
func (m *Monitor) haltDependentStrategiesLocked(pluginID string) {
	strategies, exists := m.pluginStrategies[pluginID]
	if !exists {
		return
	}
	for _, strategyID := range strategies {
		m.pausedStrategies[strategyID] = true
	}
}

// IsStrategyPaused returns whether the given strategy has been halted by a tripped circuit breaker.
func (m *Monitor) IsStrategyPaused(strategyID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pausedStrategies[strategyID]
}
