package abi

import (
	"encoding/json"
	"testing"
)

func TestMarketStateJSON(t *testing.T) {
	ms := MarketState{
		Symbol:    "BTC-USD",
		Price:     50000.5,
		Volume:    10.2,
		Timestamp: 1600000000,
	}

	b, err := json.Marshal(ms)
	if err != nil {
		t.Fatalf("Failed to marshal MarketState: %v", err)
	}

	var decoded MarketState
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal MarketState: %v", err)
	}

	if decoded.Symbol != ms.Symbol || decoded.Price != ms.Price {
		t.Errorf("Mismatch after unmarshal. Expected %v, got %v", ms, decoded)
	}
}

func TestTradeSignalJSON(t *testing.T) {
	ts := TradeSignal{
		Action:     "BUY",
		Confidence: 0.95,
		Size: 1.5, Price: 1.0,
	}

	b, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("Failed to marshal TradeSignal: %v", err)
	}

	var decoded TradeSignal
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TradeSignal: %v", err)
	}

	if decoded.Action != ts.Action || decoded.Confidence != ts.Confidence {
		t.Errorf("Mismatch after unmarshal. Expected %v, got %v", ts, decoded)
	}
}

func TestManifestJSON(t *testing.T) {
	manifest := Manifest{
		Name:        "MomentumStrategy",
		Version:     "1.0.0",
		Description: "A simple momentum strategy",
		Author:      "Alice",
	}

	b, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal Manifest: %v", err)
	}

	var decoded Manifest
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Manifest: %v", err)
	}

	if decoded.Name != manifest.Name || decoded.Version != manifest.Version {
		t.Errorf("Mismatch after unmarshal. Expected %v, got %v", manifest, decoded)
	}
}
