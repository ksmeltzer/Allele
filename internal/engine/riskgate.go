package engine

import (
	"allele/internal/abi"
)

type RiskGate struct {
	MaxCapital        float64 // Configurable global capital cap (defaults to 100.0 if not set)
	BanAsymmetricRisk bool    // Configurable toggle for highly asymmetric downside risk (default true)
}

func NewRiskGate(maxCapital float64, banAsymmetricRisk bool) *RiskGate {
	if maxCapital <= 0 {
		maxCapital = 100.0 // Default fallback
	}
	return &RiskGate{
		MaxCapital:        maxCapital,
		BanAsymmetricRisk: banAsymmetricRisk,
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
		
		// In prediction markets, buying an outcome > $0.90 to make < $0.10 is asymmetric downside risk.
		// If the ban is enabled, we reject these trades.
		if rg.BanAsymmetricRisk && sig.Price > 0.90 {
			rejectionReasons = append(rejectionReasons, "Signal rejected: Asymmetric Risk Ban active (Price > 0.90)")
			continue
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
