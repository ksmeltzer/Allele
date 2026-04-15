package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"allele/internal/core"
	"allele/internal/execution"
	"allele/internal/polymarket"
	"allele/internal/storage"
	"github.com/ethereum/go-ethereum/common"
)

type PolymarketExchange struct {
	wsClient   *polymarket.WsClient
	restClient *execution.Client
	eventBus   *core.EventBus
}

func NewPolymarketExchange(ws *polymarket.WsClient, rest *execution.Client, eventBus *core.EventBus) *PolymarketExchange {
	p := &PolymarketExchange{
		wsClient:   ws,
		restClient: rest,
		eventBus:   eventBus,
	}

	if eventBus != nil {
		go p.listenForConfig()
		go p.validateCredentials(time.Second * 2)
	}

	return p
}

func (p *PolymarketExchange) validateCredentials(delay time.Duration) {
	if delay > 0 {
		time.Sleep(delay)
	}

	key, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLY_API_KEY")
	secret, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLY_API_SECRET")

	// If missing API keys, but the user provided Wallet Address & Private Key in this plugin's config, auto-generate them.
	if key == "" || secret == "" {
		walletAddress, _ := storage.GetPluginConfig("allele-exchange-polymarket", "WALLET_ADDRESS")
		walletPrivKey, _ := storage.GetPluginConfig("allele-exchange-polymarket", "WALLET_PRIVATE_KEY")

		if walletAddress != "" && walletPrivKey != "" {
			p.eventBus.Publish(core.Event{
				Type: core.SystemAlertEvent,
				Payload: map[string]interface{}{
					"source":  "allele-exchange-polymarket",
					"level":   "warning",
					"message": "Auto-generating Polymarket API keys using provided wallet credentials...",
				},
			})

			newKey, newSecret, newPassphrase, err := execution.GenerateKeysFromWallet(walletAddress, walletPrivKey)
			if err == nil && newKey != "" {
				// Store the newly generated keys
				storage.SetPluginConfig("allele-exchange-polymarket", "POLY_API_KEY", newKey)
				storage.SetPluginConfig("allele-exchange-polymarket", "POLY_API_SECRET", newSecret)
				storage.SetPluginConfig("allele-exchange-polymarket", "POLY_API_PASSPHRASE", newPassphrase)

				// Update the local instance variables so the next ping works
				p.restClient = execution.NewClient(newKey, newSecret, newPassphrase)
				key = newKey
				secret = newSecret
			} else {
				p.eventBus.Publish(core.Event{
					Type: core.SystemAlertEvent,
					Payload: map[string]interface{}{
						"source":  "allele-exchange-polymarket",
						"level":   "error",
						"message": fmt.Sprintf("Failed to auto-generate API keys: %v", err),
					},
				})
				return
			}
		}
	}

	if key == "" || secret == "" {
		p.eventBus.Publish(core.Event{
			Type: core.SystemAlertEvent,
			Payload: map[string]interface{}{
				"source":  "allele-exchange-polymarket",
				"level":   "warning",
				"message": "Missing Polymarket credentials. Orders will fail. Please configure the plugin.",
			},
		})
	} else if len(key) < 10 { // naive junk check
		p.eventBus.Publish(core.Event{
			Type: core.SystemAlertEvent,
			Payload: map[string]interface{}{
				"source":  "allele-exchange-polymarket",
				"level":   "error",
				"message": "Invalid Polymarket API Key format detected. Connection will fail.",
			},
		})
	} else {
		// Ping the Polymarket API to verify the credentials actually work
		err := p.restClient.PingAuth()
		if err != nil {
			p.eventBus.Publish(core.Event{
				Type: core.SystemAlertEvent,
				Payload: map[string]interface{}{
					"source":  "allele-exchange-polymarket",
					"level":   "error",
					"message": fmt.Sprintf("Polymarket authentication failed: %v", err),
				},
			})
		} else {
			p.eventBus.Publish(core.Event{
				Type: core.SystemAlertEvent,
				Payload: map[string]interface{}{
					"source":  "allele-exchange-polymarket",
					"level":   "info",
					"message": "Polymarket credentials verified and connected.",
				},
			})
		}
	}
}

func (p *PolymarketExchange) listenForConfig() {
	ch := p.eventBus.Subscribe(core.ConfigUpdatedEvent)
	for event := range ch {
		payload, ok := event.Payload.(map[string]string)
		if !ok || payload["plugin_name"] != "allele-exchange-polymarket" {
			continue
		}

		log.Println("PolymarketExchange: Config updated, re-initializing credentials...")

		key, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLY_API_KEY")
		secret, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLY_API_SECRET")
		passphrase, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLY_API_PASSPHRASE")

		p.restClient = execution.NewClient(key, secret, passphrase)

		p.validateCredentials(0)
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
