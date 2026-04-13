package engine

import (
	"allele/internal/abi"
	"testing"
)

func TestRiskGate_ValidateSignals(t *testing.T) {
	rg := NewRiskGate(100.0, true) // Test with 100 max cap and asymmetric ban ON

	signals := []abi.TradeSignal{
		{Action: "BUY", Size: 10, Price: 0.5},   // cost = 5, valid
		{Action: "BUY", Size: 200, Price: 0.6},  // cost = 120, exceeds global cap
		{Action: "BUY", Size: -5, Price: 0.10},  // Size <= 0, invalid
		{Action: "BUY", Size: 100, Price: 0.8},  // cost = 80, exceeds available capital (if available is 70)
		{Action: "BUY", Size: 0, Price: 0.10},   // Size <= 0, invalid
		{Action: "BUY", Size: 10, Price: 0.95},  // Price > 0.90, asymmetric risk ban active
	}

	availableCapital := 70.0
	valid, reasons, err := rg.ValidateSignals(signals, availableCapital)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(valid) != 1 {
		t.Fatalf("expected 1 valid signal, got %d", len(valid))
	}

	if valid[0].Size != 10 || valid[0].Price != 0.5 {
		t.Errorf("expected valid signal to be Size 10 Price 0.5, got Size %f Price %f", valid[0].Size, valid[0].Price)
	}

	expectedReasons := 5
	if len(reasons) != expectedReasons {
		t.Errorf("expected %d rejection reasons, got %d", expectedReasons, len(reasons))
	}
}

func TestRiskGate_CustomCap(t *testing.T) {
	rg := NewRiskGate(500.0, false) // Test with 500 max cap and asymmetric ban OFF

	signals := []abi.TradeSignal{
		{Action: "BUY", Size: 400, Price: 1.0}, // cost = 400, valid, Price = 1.0 would normally trigger asymmetric risk but it's OFF
		{Action: "BUY", Size: 600, Price: 1.0}, // cost = 600, exceeds custom global cap
	}

	availableCapital := 1000.0
	valid, _, err := rg.ValidateSignals(signals, availableCapital)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(valid) != 1 {
		t.Fatalf("expected 1 valid signal, got %d", len(valid))
	}
}
