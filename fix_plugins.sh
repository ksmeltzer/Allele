#!/bin/bash
cat << 'PLUGIN' > plugins/allele-strategy-completeness-go/main.go
//go:build wasm

package main

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

//go:wasmexport Manifest
func Manifest() uint64 {
	manifestJSON := `{"name": "allele-strategy-completeness-go", "version": "v1.0.0", "description": "Completeness Arbitrage Engine", "author": "Allele Org", "dependencies": [{"name": "allele-exchange-polymarket", "type": "exchange", "version": ">=v1.0.0", "url": "https://github.com/ksmeltzer/allele-exchange-polymarket/releases/download/v1.0.0/plugin.wasm"}], "config": [{"key": "PROFIT_MARGIN", "title": "Minimum Profit Margin", "type": "number", "description": "Required risk-free profit percentage", "explanation": "The guaranteed risk-free profit percentage required to trigger a completeness arbitrage (simultaneously buying all mutually exclusive outcomes in a market). This calculation currently does not account for Polygon network gas fees.", "defaultValue": "0.01", "required": true}]}`
	outBytes := []byte(manifestJSON)
	outPtr := uint32(uintptr(unsafe.Pointer(&outBytes[0])))
	outLen := uint32(len(outBytes))
	return (uint64(outPtr) << 32) | uint64(outLen)
}

type MarketState struct {
	AssetPrices map[string]float64 `json:"AssetPrices"`
}

type TradeSignal struct {
	Action     string                 `json:"action"`
	Confidence float64                `json:"confidence"`
	Size       float64                `json:"size,omitempty"`
	Price      float64                `json:"price,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

//go:wasmexport malloc
func malloc(size uint32) uint32 {
	buf := make([]byte, size)
	return uint32(uintptr(unsafe.Pointer(&buf[0])))
}

//go:wasmexport free
func free(ptr uint32, size uint32) {}

func readMemory(ptr uint32, length uint32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), length)
}

func writeMemory(data []byte) (uint32, uint32) {
	ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
	return ptr, uint32(len(data))
}

//go:wasmexport Evaluate
func Evaluate(statePtr uint32, stateLength uint32) uint64 {
	stateBytes := readMemory(statePtr, stateLength)

	var state MarketState
	if err := json.Unmarshal(stateBytes, &state); err != nil {
		return 0
	}

	var sum float64
	for _, price := range state.AssetPrices {
		sum += price
	}

	signals := []TradeSignal{}

	// Fee is mock for testing
	fee := 0.02
	if sum > 0 && sum+fee < 1.0 {
		for tokenID, price := range state.AssetPrices {
			signals = append(signals, TradeSignal{
				Action:     "BUY",
				Confidence: 1.0,
				Size:       1.0,
				Price:      price,
				Metadata: map[string]interface{}{
					"asset_id": tokenID,
					"reason":   fmt.Sprintf("Completeness Arbitrage. Sum=%.2f", sum),
				},
			})
		}
	} else if len(state.AssetPrices) > 0 {
        // Emit a mock HOLD signal just to prove the engine evaluates state
        // and sends signals to the UI!
        for tokenID, price := range state.AssetPrices {
			signals = append(signals, TradeSignal{
				Action:     "HOLD",
				Confidence: 0.5,
				Size:       0.0,
				Price:      price,
				Metadata: map[string]interface{}{
					"asset_id": tokenID,
					"reason":   fmt.Sprintf("Market observed, no arbitrage. Sum=%.2f", sum),
				},
			})
            break
		}
    }

	outBytes, _ := json.Marshal(signals)
	outPtr, outLen := writeMemory(outBytes)

	return (uint64(outPtr) << 32) | uint64(outLen)
}

func main() { select {} }
PLUGIN
