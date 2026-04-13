package engine

import (
	"allele/internal/abi"
)

type RiskGate struct{}

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
		if cost > 100.0 {
			rejectionReasons = append(rejectionReasons, "Signal rejected: Exceeds $100 global cap")
			continue // Enforce the $100 global cap
		}
		validSignals = append(validSignals, sig)
	}
	return validSignals, rejectionReasons, nil
}
