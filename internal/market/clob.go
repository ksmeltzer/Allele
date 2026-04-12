package market

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

// ============================================================================
// Polymarket WebSocket Event Types
// ============================================================================

// BaseEvent is used to sniff the event_type from the raw JSON payload.
type BaseEvent struct {
	EventType string `json:"event_type"`
}

type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type BookEvent struct {
	EventType string       `json:"event_type"`
	AssetID   string       `json:"asset_id"`
	Market    string       `json:"market"`
	Bids      []PriceLevel `json:"bids"`
	Asks      []PriceLevel `json:"asks"`
}

type PriceChange struct {
	AssetID string `json:"asset_id"`
	Side    string `json:"side"` // "BUY" or "SELL"
	Price   string `json:"price"`
	Size    string `json:"size"`
}

type PriceChangeEvent struct {
	EventType    string        `json:"event_type"`
	Market       string        `json:"market"`
	PriceChanges []PriceChange `json:"price_changes"`
}

// ============================================================================
// Central Limit Order Book (CLOB) State Manager
// ============================================================================

// OrderBook represents the current liquidity state for a single asset.
type OrderBook struct {
	mu      sync.RWMutex
	AssetID string
	// Price (string) -> Size (decimal)
	// We use string keys for O(1) exact matching, parsing to decimal when sorting.
	Bids map[string]decimal.Decimal
	Asks map[string]decimal.Decimal
}

// TopOfBook returns the best bid and best ask, alongside their sizes.
func (ob *OrderBook) TopOfBook() (bestBid, bidSize, bestAsk, askSize decimal.Decimal, err error) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	// Extract and sort Bids (Descending - Highest Bid First)
	var bidPrices []decimal.Decimal
	for pStr := range ob.Bids {
		p, _ := decimal.NewFromString(pStr)
		bidPrices = append(bidPrices, p)
	}
	sort.Slice(bidPrices, func(i, j int) bool {
		return bidPrices[i].GreaterThan(bidPrices[j])
	})

	if len(bidPrices) > 0 {
		bestBid = bidPrices[0]
		bidSize = ob.Bids[bestBid.String()]
	}

	// Extract and sort Asks (Ascending - Lowest Ask First)
	var askPrices []decimal.Decimal
	for pStr := range ob.Asks {
		p, _ := decimal.NewFromString(pStr)
		askPrices = append(askPrices, p)
	}
	sort.Slice(askPrices, func(i, j int) bool {
		return askPrices[i].LessThan(askPrices[j])
	})

	if len(askPrices) > 0 {
		bestAsk = askPrices[0]
		askSize = ob.Asks[bestAsk.String()]
	}

	if len(bidPrices) == 0 || len(askPrices) == 0 {
		return bestBid, bidSize, bestAsk, askSize, fmt.Errorf("orderbook is empty on one or both sides")
	}

	return bestBid, bidSize, bestAsk, askSize, nil
}

// CLOBManager holds the state of all order books the engine is tracking.
type CLOBManager struct {
	mu    sync.RWMutex
	Books map[string]*OrderBook // AssetID -> OrderBook
}

// NewCLOBManager initializes a new state manager.
func NewCLOBManager() *CLOBManager {
	return &CLOBManager{
		Books: make(map[string]*OrderBook),
	}
}

// GetBook returns a thread-safe pointer to an asset's order book.
func (cm *CLOBManager) GetBook(assetID string) *OrderBook {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	book, exists := cm.Books[assetID]
	if !exists {
		book = &OrderBook{
			AssetID: assetID,
			Bids:    make(map[string]decimal.Decimal),
			Asks:    make(map[string]decimal.Decimal),
		}
		cm.Books[assetID] = book
	}
	return book
}

// ProcessWebSocketMessage routes raw JSON from the Polymarket WS into the state machine.
func (cm *CLOBManager) ProcessWebSocketMessage(raw []byte) error {
	// First, sniff the array. Polymarket wraps events in arrays.
	var events []json.RawMessage
	if err := json.Unmarshal(raw, &events); err != nil {
		// Sometimes it's a single object, not an array.
		events = []json.RawMessage{raw}
	}

	for _, rawEvent := range events {
		var base BaseEvent
		if err := json.Unmarshal(rawEvent, &base); err != nil {
			continue // Skip unparseable events
		}

		switch base.EventType {
		case "book":
			var snap BookEvent
			if err := json.Unmarshal(rawEvent, &snap); err == nil {
				cm.applySnapshot(snap)
			}
		case "price_change":
			var delta PriceChangeEvent
			if err := json.Unmarshal(rawEvent, &delta); err == nil {
				cm.applyDelta(delta)
			}
			// Ignore "last_trade_price" and others for now, as they don't affect CLOB depth directly.
		}
	}
	return nil
}

func (cm *CLOBManager) applySnapshot(snap BookEvent) {
	book := cm.GetBook(snap.AssetID)

	book.mu.Lock()
	defer book.mu.Unlock()

	// Clear existing state
	book.Bids = make(map[string]decimal.Decimal)
	book.Asks = make(map[string]decimal.Decimal)

	for _, b := range snap.Bids {
		size, _ := decimal.NewFromString(b.Size)
		book.Bids[b.Price] = size
	}
	for _, a := range snap.Asks {
		size, _ := decimal.NewFromString(a.Size)
		book.Asks[a.Price] = size
	}

	log.Printf("[CLOB Snapshot] Asset: %s | Bids: %d | Asks: %d", snap.AssetID, len(book.Bids), len(book.Asks))
}

func (cm *CLOBManager) applyDelta(delta PriceChangeEvent) {
	// Group changes by AssetID
	changesByAsset := make(map[string][]PriceChange)
	for _, pc := range delta.PriceChanges {
		changesByAsset[pc.AssetID] = append(changesByAsset[pc.AssetID], pc)
	}

	for assetID, changes := range changesByAsset {
		book := cm.GetBook(assetID)

		book.mu.Lock()
		for _, pc := range changes {
			size, _ := decimal.NewFromString(pc.Size)

			if pc.Side == "BUY" {
				if size.IsZero() {
					delete(book.Bids, pc.Price)
				} else {
					book.Bids[pc.Price] = size
				}
			} else if pc.Side == "SELL" {
				if size.IsZero() {
					delete(book.Asks, pc.Price)
				} else {
					book.Asks[pc.Price] = size
				}
			}
		}
		book.mu.Unlock()
	}
}
