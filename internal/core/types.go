package core

import "time"

type ActionType string

const (
	BUY  ActionType = "BUY"
	SELL ActionType = "SELL"
	HOLD ActionType = "HOLD"
)

type OrderbookState struct {
	Bids      map[string]float64
	Asks      map[string]float64
	TopBid    float64
	TopAsk    float64
	TotalSize float64
}

type MarketState struct {
	AssetPrices map[string]float64
}

type Action struct {
	MarketID   string
	AssetID    string
	AssetName  string
	Side       ActionType
	Price      float64
	Size       float64
	StrategyID string
}

type NormalizedTick struct {
	MarketID  string
	AssetID   string
	AssetName string
	IsBid     bool
	Price     float64
	Size      float64
	Timestamp time.Time
}
