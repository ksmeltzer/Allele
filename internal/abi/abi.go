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
	Action     string                 `json:"action"`     // e.g., "BUY", "SELL", "HOLD"
	Confidence float64                `json:"confidence"` // 0.0 to 1.0
	Size       float64                `json:"size,omitempty"`
	Price      float64                `json:"price,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // Arbitrary telemetry to render on the UI
}

// ConfigField defines a parameter that the plugin requires the user to configure via the UI.
type ConfigField struct {
	Key          string   `json:"key"`
	Title        string   `json:"title"` // New: UI Title
	Type         string   `json:"type"`  // "string", "number", "boolean", "secret", "enum"
	Description  string   `json:"description"`
	Explanation  string   `json:"explanation,omitempty"`  // Deep dive explanation for tooltips
	DefaultValue string   `json:"defaultValue,omitempty"` // Default value
	Required     bool     `json:"required"`
	Options      []string `json:"options,omitempty"`     // Allowed values if type == "enum"
	MultiSelect  bool     `json:"multiSelect,omitempty"` // If true and type == "enum", allow multiple values
}

// Dependency defines another plugin that this plugin requires to function.
type Dependency struct {
	Name    string `json:"name"`
	Type    string `json:"type"`    // "exchange", "sensor", "strategy"
	Version string `json:"version"` // Semver requirement (e.g. ">=v1.0.0")
	Url     string `json:"url"`     // Download URL for the required WASM module
}

// Manifest represents the metadata for a Tri-Plugin WASM module.
// Every plugin must export a "manifest" function that returns this structure
// serialized to JSON so the engine can determine its dependencies and config UI.
type Manifest struct {
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Description  string        `json:"description"`
	Author       string        `json:"author"`
	Dependencies []Dependency  `json:"dependencies"`
	Config       []ConfigField `json:"config"`
}
