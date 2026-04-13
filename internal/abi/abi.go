package abi

// MarketState represents the current state of the market passed to the WASM plugin.
type MarketState struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
}

// TradeSignal represents the decision output from the WASM plugin.
type TradeSignal struct {
	Action     string  `json:"action"`     // e.g., "BUY", "SELL", "HOLD"
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
	Size       float64 `json:"size,omitempty"`
	Price      float64 `json:"price,omitempty"`
}

// Manifest represents the metadata for a Tri-Plugin WASM module.
type Manifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
}
