package engine

import (
	"allele/internal/abi"
	"testing"
)

func TestRiskGate_ValidateSignals(t *testing.T) {
	rg := NewRiskGate(100.0) // Test with 100 max cap

	signals := []abi.TradeSignal{
		{Action: "BUY", Size: 10, Price: 5},  // cost = 50, valid
		{Action: "BUY", Size: 20, Price: 6},  // cost = 120, exceeds global cap
		{Action: "BUY", Size: -5, Price: 10}, // Size <= 0, invalid
		{Action: "BUY", Size: 10, Price: 8},  // cost = 80, exceeds available capital (if available is 70)
		{Action: "BUY", Size: 0, Price: 10},  // Size <= 0, invalid
	}

	availableCapital := 70.0
	valid, reasons, err := rg.ValidateSignals(signals, availableCapital)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(valid) != 1 {
		t.Fatalf("expected 1 valid signal, got %d", len(valid))
	}

	if valid[0].Size != 10 || valid[0].Price != 5 {
		t.Errorf("expected valid signal to be Size 10 Price 5, got Size %f Price %f", valid[0].Size, valid[0].Price)
	}

	expectedReasons := 4
	if len(reasons) != expectedReasons {
		t.Errorf("expected %d rejection reasons, got %d", expectedReasons, len(reasons))
	}
}

func TestRiskGate_CustomCap(t *testing.T) {
	rg := NewRiskGate(500.0) // Test with 500 max cap

	signals := []abi.TradeSignal{
		{Action: "BUY", Size: 40, Price: 10}, // cost = 400, valid
		{Action: "BUY", Size: 60, Price: 10}, // cost = 600, exceeds custom global cap
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
