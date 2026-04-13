package core

import "context"

type IWallet interface {
	SignMessage(payload []byte) ([]byte, error)
	Address() string
}

type IExchange interface {
	ConnectStream(ctx context.Context, tickChan chan<- NormalizedTick) error
	SubmitOrder(ctx context.Context, wallet IWallet, action Action) error
}

type IStrategy interface {
	ID() string
	Evaluate(state *MarketState) []Action
	GetDNA() map[string]float64
}

type IEngine interface {
	RegisterExchange(e IExchange)
	RegisterWallet(w IWallet)
	RegisterStrategy(s IStrategy)
	Start(ctx context.Context)
}
