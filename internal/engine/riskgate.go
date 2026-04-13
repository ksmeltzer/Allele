package engine

import (
	"allele/internal/abi"
)

type RiskGate struct {
	MaxCapital float64 // Configurable global capital cap (defaults to 100.0 if not set)
}

func NewRiskGate(maxCapital float64) *RiskGate {
	if maxCapital <= 0 {
		maxCapital = 100.0 // Default fallback
	}
	return &RiskGate{
		MaxCapital: maxCapital,
	}
}

func (rg *RiskGate) ValidateSignals(signals []abi.TradeSignal, availableCapital float64) ([]abi.TradeSignal, []string, error) {
	var validSignals []abi.TradeSignal
	var rejectionReasons []string

	for _, sig := range signals {
		if sig.Size <= 0 {
			rejectionReasons = append(rejectionReasons, "Signal rejected: Size <= 0")
			continue // Enforce Size > 0
		}
		cost := sig.Size * sig.Price
		if cost > availableCapital {
			rejectionReasons = append(rejectionReasons, "Signal rejected: Exceeds available capital")
			continue // Reject if exceeds available capital
		}
		if rg.MaxCapital > 0 && cost > rg.MaxCapital {
			rejectionReasons = append(rejectionReasons, "Signal rejected: Exceeds global capital cap")
			continue // Enforce the configurable global cap
		}
		validSignals = append(validSignals, sig)
	}
	return validSignals, rejectionReasons, nil
}
