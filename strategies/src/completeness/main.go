//go:build wasm
package main

import (
	"encoding/json"
	"fmt"
)

// This is a minimal, illustrative WASM Plugin for Allele.
// The engine (host) will invoke the Evaluate function, passing memory pointers.

// MarketState is the structure we expect to receive from the host
type MarketState struct {
	MarketID string             `json:"market_id"`
	Tokens   map[string]float64 `json:"tokens"` // token_id -> best_ask_price
}

// Signal is what we return to the host
type Signal struct {
	Asset      string  `json:"asset"`
	Action     string  `json:"action"`     // BUY or SELL
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
	Reason     string  `json:"reason"`
}

// These functions must be provided by the host environment (Go/Wazero)
// so the WASM module can read and write strings.
//
//go:wasmimport env read_memory
func _readMemory(ptr uint32, length uint32) uint32

//go:wasmimport env write_memory
func _writeMemory(ptr uint32, length uint32) uint32

func readMemory(ptr uint32, length uint32) []byte { return []byte("{}") }
func writeMemory(data []byte) (uint32, uint32)    { return 0, 0 }

//export Evaluate
func Evaluate(statePtr uint32, stateLength uint32) (uint32, uint32) {
	// 1. Read the JSON market state from the host's memory
	stateBytes := readMemory(statePtr, stateLength)

	var state MarketState
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		// In a real plugin, handle errors or return an empty signal array
		return 0, 0
	}

	// 2. Execute Completeness Arbitrage Logic
	// Sum the prices of all mutually exclusive tokens
	var sum float64
	for _, price := range state.Tokens {
		sum += price
	}

	signals := []Signal{}

	// If the sum is less than 1.00 (discounting a mock fee), it's an arbitrage
	// The exact fee and min profit margin would be passed in as 'Genes'
	// from the genetic algorithm, but they are hardcoded for this simple example.
	fee := 0.02
	if sum+fee < 1.0 {
		for tokenID := range state.Tokens {
			signals = append(signals, Signal{
				Asset:      tokenID,
				Action:     "BUY",
				Confidence: 1.0,
				Reason:     fmt.Sprintf("Completeness Arbitrage. Sum=%.2f", sum),
			})
		}
	}

	// 3. Serialize the signals back to JSON and write them to the host's memory
	outBytes, _ := json.Marshal(signals)
	outPtr, outLen := writeMemory(outBytes)

	// 4. Return the pointer to the signals so the Go host can execute them
	return outPtr, outLen
}

// The main function is required for TinyGo/Go WASM compilation,
// but the entry point for the engine is the exported Evaluate function above.
func main() {}
