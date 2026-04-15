package strategy

import (
	"context"
	"encoding/json"
	"log"

	"allele/internal/abi"
	"allele/internal/core"
	"allele/internal/loader"
)

type WasmStrategy struct {
	id     string
	module *loader.WasmModule
	dna    map[string]float64
}

func NewWasmStrategy(id string, mod *loader.WasmModule) *WasmStrategy {
	return &WasmStrategy{
		id:     id,
		module: mod,
		dna:    make(map[string]float64),
	}
}

func (w *WasmStrategy) ID() string {
	return w.id
}

func (w *WasmStrategy) GetDNA() map[string]float64 {
	return w.dna
}

func (w *WasmStrategy) Evaluate(state *core.MarketState) []core.Action {
	ctx := context.Background()

	stateBytes, err := json.Marshal(state)
	if err != nil {
		log.Printf("WasmStrategy %s: failed to marshal state: %v", w.id, err)
		return nil
	}

	ptr, length, err := abi.WriteToMemory(ctx, w.module.Module, stateBytes)
	if err != nil {
		log.Printf("WasmStrategy %s: WriteToMemory failed: %v", w.id, err)
		return nil
	}

	evalFn := w.module.Module.ExportedFunction("Evaluate")
	if evalFn == nil {
		log.Printf("WasmStrategy %s: module does not export 'Evaluate'", w.id)
		return nil
	}

	res, err := evalFn.Call(ctx, uint64(ptr), uint64(length))
	if err != nil {
		log.Printf("WasmStrategy %s: Evaluate call failed: %v", w.id, err)
		return nil
	}

	if len(res) == 0 {
		return nil
	}

	val := res[0]
	if val == 0 {
		return nil
	}

	outPtr := uint32(val >> 32)
	outLen := uint32(val & 0xFFFFFFFF)

	outBytes, err := abi.ReadFromMemory(ctx, w.module.Module, outPtr, outLen)
	if err != nil {
		log.Printf("WasmStrategy %s: failed to read result: %v", w.id, err)
		return nil
	}

	var signals []abi.TradeSignal
	if err := json.Unmarshal(outBytes, &signals); err != nil {
		log.Printf("WasmStrategy %s: failed to unmarshal signals: %v", w.id, err)
		return nil
	}

	var actions []core.Action
	for _, sig := range signals {
		side := core.BUY
		if sig.Action == "SELL" {
			side = core.SELL
		} else if sig.Action == "HOLD" {
			side = core.HOLD
		}

		var assetID string
		if sig.Metadata != nil {
			if id, ok := sig.Metadata["asset_id"].(string); ok {
				assetID = id
			}
		}
		
		actions = append(actions, core.Action{
			StrategyID: w.id,
			MarketID:   "polymarket", // Extract from metadata or default to polymarket
			AssetID:    assetID,
			Side:       side,
			Price:      sig.Price,
			Size:       sig.Size,
		})
	}

	return actions
}
