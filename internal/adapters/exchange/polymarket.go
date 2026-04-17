package exchange

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"allele/internal/adapters/wallet"
	"allele/internal/core"
	"allele/internal/execution"
	"allele/internal/polymarket"
	"allele/internal/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
		go p.healthCheckLoop()
	}

	return p
}

func (p *PolymarketExchange) checkAndApprove(rpcURL, network, privKeyHex string) {
	if rpcURL == "" || privKeyHex == "" {
		return
	}

	pkBytes := common.FromHex(privKeyHex)
	if len(pkBytes) == 0 {
		return
	}
	pk, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		return
	}

	rpcManager, err := wallet.NewRPCManager(rpcURL, network, pk)
	if err != nil {
		log.Printf("PolymarketExchange: Failed to connect to RPC for balances: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rpcManager.CheckAndApproveUSDC(ctx); err != nil {
		log.Printf("PolymarketExchange: CheckAndApproveUSDC failed: %v", err)
	} else {
		log.Printf("PolymarketExchange: USDC Allowance verified successfully.")
	}

	matic, usdc, err := rpcManager.GetBalances(ctx)
	if err == nil {
		// Calculate true floats (MATIC has 18 decimals, USDC has 6)
		maticFloat := new(big.Float).SetInt(matic)
		maticFloat.Quo(maticFloat, big.NewFloat(1e18))
		mFloat, _ := maticFloat.Float64()

		usdcFloat := new(big.Float).SetInt(usdc)
		usdcFloat.Quo(usdcFloat, big.NewFloat(1e6))
		uFloat, _ := usdcFloat.Float64()

		p.eventBus.Publish(core.Event{
			Type: "wallet_balance",
			Payload: map[string]interface{}{
				"address": crypto.PubkeyToAddress(pk.PublicKey).Hex(),
				"network": network,
				"matic":   mFloat,
				"usdc":    uFloat,
			},
		})
	} else {
		log.Printf("PolymarketExchange: GetBalances failed: %v", err)
	}
}

func (p *PolymarketExchange) validateCredentials(delay time.Duration) {
	if delay > 0 {
		time.Sleep(delay)
	}

	key, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLY_API_KEY")
	secret, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLY_API_SECRET")

	walletAddress, _ := storage.GetPluginConfig("allele-exchange-polymarket", "WALLET_ADDRESS")
	walletPrivKey, _ := storage.GetPluginConfig("allele-exchange-polymarket", "WALLET_PRIVATE_KEY")
	rpcURL, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLYGON_RPC_URL")
	network, _ := storage.GetPluginConfig("allele-exchange-polymarket", "NETWORK")

	// Auto-heal: Replace dead default RPCs with the working Amoy one
	if rpcURL == "" || rpcURL == "https://polygon-rpc.com" {
		rpcURL = "https://rpc-amoy.polygon.technology"
		storage.SetPluginConfig("allele-exchange-polymarket", "POLYGON_RPC_URL", rpcURL, false)
	}

	// Auto-heal: Default to Simulation Mode (Amoy Testnet) if unconfigured
	if network == "" {
		network = "Polygon Amoy Testnet"
		storage.SetPluginConfig("allele-exchange-polymarket", "NETWORK", network, false)
	}

	// Auto-generate crypto wallet if none exists
	if walletPrivKey == "" {
		privateKey, err := crypto.GenerateKey()
		if err == nil {
			privateKeyBytes := crypto.FromECDSA(privateKey)
			walletPrivKey = common.Bytes2Hex(privateKeyBytes)
			publicKey := privateKey.Public()
			if publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey); ok {
				walletAddress = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
				storage.SetPluginConfig("allele-exchange-polymarket", "WALLET_PRIVATE_KEY", walletPrivKey, true)
				storage.SetPluginConfig("allele-exchange-polymarket", "WALLET_ADDRESS", walletAddress, false)

				if p.eventBus != nil {
					p.eventBus.Publish(core.Event{
						Type: core.SystemAlertEvent,
						Payload: map[string]interface{}{
							"source":  "allele-exchange-polymarket",
							"level":   "info",
							"message": "Auto-generated new Simulation Wallet for Polymarket.",
						},
					})
				}
			}
		}
	}

	// If missing API keys, but we now have Wallet Address & Private Key, auto-generate them.
	if key == "" || secret == "" {
		if walletAddress != "" && walletPrivKey != "" {
			go p.checkAndApprove(rpcURL, network, walletPrivKey)

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
				// Store the newly generated keys under the hood without exposing them in the UI config manifest
				storage.SetPluginConfig("allele-exchange-polymarket", "POLY_API_KEY", newKey, true)
				storage.SetPluginConfig("allele-exchange-polymarket", "POLY_API_SECRET", newSecret, true)
				storage.SetPluginConfig("allele-exchange-polymarket", "POLY_API_PASSPHRASE", newPassphrase, true)

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
		} else {
			p.eventBus.Publish(core.Event{
				Type: core.SystemAlertEvent,
				Payload: map[string]interface{}{
					"source":  "allele-exchange-polymarket",
					"level":   "warning",
					"message": "Missing Polygon Wallet credentials. Please configure the plugin.",
				},
			})
			return
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
			walletPrivKey, _ := storage.GetPluginConfig("allele-exchange-polymarket", "WALLET_PRIVATE_KEY")
			rpcURL, _ := storage.GetPluginConfig("allele-exchange-polymarket", "POLYGON_RPC_URL")
			network, _ := storage.GetPluginConfig("allele-exchange-polymarket", "NETWORK")

			if walletPrivKey != "" {
				go p.checkAndApprove(rpcURL, network, walletPrivKey)
			}

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

type BookEvent struct {
	MarketID  string `json:"market"`
	AssetID   string `json:"asset_id"`
	EventType string `json:"event_type"`
	Bids      []struct {
		Price string `json:"price"`
		Size  string `json:"size"`
	} `json:"bids"`
	Asks []struct {
		Price string `json:"price"`
		Size  string `json:"size"`
	} `json:"asks"`
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
				var events []BookEvent
				if err := json.Unmarshal(msg, &events); err == nil {
					for _, ev := range events {
						if ev.EventType == "book" && ev.AssetID != "" {
							// Parse top bid
							if len(ev.Bids) > 0 {
								price, _ := strconv.ParseFloat(ev.Bids[0].Price, 64)
								size, _ := strconv.ParseFloat(ev.Bids[0].Size, 64)
								tickChan <- core.NormalizedTick{
									MarketID:  ev.MarketID,
									AssetID:   ev.AssetID,
									AssetName: polymarket.GetAssetMetadata(ev.AssetID),
									IsBid:     true,
									Price:     price,
									Size:      size,
									Timestamp: time.Now(),
								}
							}
							// Parse top ask
							if len(ev.Asks) > 0 {
								price, _ := strconv.ParseFloat(ev.Asks[0].Price, 64)
								size, _ := strconv.ParseFloat(ev.Asks[0].Size, 64)
								tickChan <- core.NormalizedTick{
									MarketID:  ev.MarketID,
									AssetID:   ev.AssetID,
									AssetName: polymarket.GetAssetMetadata(ev.AssetID),
									IsBid:     false,
									Price:     price,
									Size:      size,
									Timestamp: time.Now(),
								}
							}
						}
					}
					continue
				}

				var singleEv BookEvent
				if err := json.Unmarshal(msg, &singleEv); err == nil && singleEv.EventType == "price_change" {
					// Fallback to old price change format if they ever send it...
				}

				// Handle price_change event
				type PriceChange struct {
					AssetID string `json:"asset_id"`
					Price   string `json:"price"`
					Size    string `json:"size"`
					Side    string `json:"side"`
				}
				type PriceChangeEventWrapper struct {
					MarketID     string        `json:"market"`
					EventType    string        `json:"event_type"`
					PriceChanges []PriceChange `json:"price_changes"`
				}

				var pce PriceChangeEventWrapper
				if err := json.Unmarshal(msg, &pce); err == nil && pce.EventType == "price_change" {
					for _, pc := range pce.PriceChanges {
						if pc.AssetID != "" {
							price, _ := strconv.ParseFloat(pc.Price, 64)
							size, _ := strconv.ParseFloat(pc.Size, 64)
							tickChan <- core.NormalizedTick{
								MarketID:  pce.MarketID,
								AssetID:   pc.AssetID,
								AssetName: polymarket.GetAssetMetadata(pc.AssetID),
								IsBid:     pc.Side == "BUY",
								Price:     price,
								Size:      size,
								Timestamp: time.Now(),
							}
						}
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
func (p *PolymarketExchange) healthCheckLoop() {
	p.validateCredentials(time.Second * 2)
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		p.validateCredentials(0)
	}
}
