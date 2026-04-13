package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"allele/internal/core"
	"allele/internal/execution"
	"allele/internal/polymarket"
	"github.com/ethereum/go-ethereum/common"
)

type PolymarketExchange struct {
	wsClient   *polymarket.WsClient
	restClient *execution.Client
}

func NewPolymarketExchange(ws *polymarket.WsClient, rest *execution.Client) *PolymarketExchange {
	return &PolymarketExchange{
		wsClient:   ws,
		restClient: rest,
	}
}

func (p *PolymarketExchange) Name() string {
	return "Polymarket"
}

type PriceChangeEvent struct {
	Event string `json:"event"`
	Data  struct {
		AssetID string  `json:"asset_id"`
		Price   float64 `json:"price"`
		Size    float64 `json:"size"`
		Side    string  `json:"side"`
	} `json:"data"`
}

func (p *PolymarketExchange) ConnectStream(ctx context.Context, tickChan chan<- core.NormalizedTick) error {
	msgChan := make(chan []byte, 100)

	// Listen pushes messages into msgChan
	go p.wsClient.Listen(ctx, msgChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgChan:
				var ev PriceChangeEvent
				if err := json.Unmarshal(msg, &ev); err != nil {
					continue
				}

				if ev.Data.AssetID != "" {
					tickChan <- core.NormalizedTick{
						MarketID:  ev.Data.AssetID,
						AssetID:   ev.Data.AssetID,
						IsBid:     ev.Data.Side == "BUY",
						Price:     ev.Data.Price,
						Size:      ev.Data.Size,
						Timestamp: time.Now(),
					}
				}
			}
		}
	}()

	return nil
}

func (p *PolymarketExchange) SubmitOrder(ctx context.Context, wallet core.IWallet, action core.Action) error {
	side := uint8(0) // BUY
	if action.Side == core.SELL {
		side = uint8(1) // SELL
	}

	addr := common.HexToAddress(wallet.Address())
	tokenID, ok := new(big.Int).SetString(action.AssetID, 10)
	if !ok {
		tokenID = big.NewInt(0)
	}

	order := execution.Order{
		Salt:          big.NewInt(time.Now().UnixNano()),
		Maker:         addr,
		Signer:        addr,
		Taker:         common.Address{},
		TokenId:       tokenID,
		MakerAmount:   big.NewInt(int64(action.Size)),
		TakerAmount:   big.NewInt(int64(action.Size)),
		Expiration:    big.NewInt(time.Now().Add(5 * time.Minute).Unix()),
		Nonce:         big.NewInt(0),
		FeeRateBps:    big.NewInt(0),
		Side:          side,
		SignatureType: 0,
	}

	hash, err := execution.HashOrder(order)
	if err != nil {
		return fmt.Errorf("failed to hash order: %w", err)
	}

	sigBytes, err := wallet.SignMessage(hash)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	sigStr := fmt.Sprintf("0x%x", sigBytes)

	return p.restClient.PlaceOrder(&order, sigStr)
}
