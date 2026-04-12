package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arbitrage/internal/config"
	"arbitrage/internal/execution"
	"arbitrage/internal/market"
	"arbitrage/internal/polymarket"
	"arbitrage/internal/strategy"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

type NewMarketEvent struct {
	EventType    string   `json:"event_type"`
	ConditionID  string   `json:"condition_id"`
	ClobTokenIDs []string `json:"clob_token_ids"`
}

func main() {
	log.Println("Starting Polymarket Data Recorder & Trading Engine...")

	// Create root context that listens for interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create channels for communication
	msgChan := make(chan []byte, 1000)

	// Initialize Historical Data Recorder
	recorder, err := market.NewRecorder("data/historical")
	if err != nil {
		log.Fatalf("Failed to initialize recorder: %v", err)
	}

	// Initialize CLOB Manager
	clobManager := market.NewCLOBManager()

	// Load config
	cfg := config.LoadConfig()

	// Initialize Execution Client
	execClient := execution.NewClient()

	var privateKey *ecdsa.PrivateKey
	if cfg.PolygonPrivateKey != "" {
		pk, err := crypto.HexToECDSA(cfg.PolygonPrivateKey)
		if err != nil {
			log.Printf("Invalid polygon private key: %v", err)
		} else {
			privateKey = pk
		}
	}

	// Initialize Strategy
	// Note: using hardcoded 2% (0.02) taker fee.
	takerFee := decimal.NewFromFloat(0.02)
	completenessArb := strategy.NewCompletenessArbitrage(clobManager, execClient, privateKey, cfg.PublicAddress, takerFee)

	// Initialize WebSocket Client
	client := polymarket.NewWsClient(polymarket.DefaultWSEndpoint)

	// Connect to WS
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Polymarket WebSocket: %v", err)
	}

	// Active Test Market IDs (From Gamma API)
	testMarketIDs := []string{
		"93592949212798121127213117304912625505836768562433217537850469496310204567695",
		"3074539347152748632858978545166555332546941892131779352477699494423276162345",
	}

	if err := client.Subscribe(testMarketIDs); err != nil {
		log.Fatalf("Failed to subscribe to initial markets: %v", err)
	}

	// Register initial markets with strategy (using dummy condition ID for testing)
	completenessArb.RegisterMarket("InitialTestMarket", testMarketIDs)

	// Start WebSocket listener
	go client.Listen(ctx, msgChan)

	// Background ticker for strategy evaluation (evaluate every 500ms)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				completenessArb.Evaluate()
			}
		}
	}()

	// Start processing messages
	go func() {
		defer recorder.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgChan:
				// 1. Record raw tick
				if string(msg) == "PONG" {
					continue // ignore heartbeats for processing
				}

				if err := recorder.WriteTick(msg); err != nil {
					log.Printf("Failed to write tick to disk: %v", err)
				}

				// 2. Dynamic Market Discovery
				// Check if the message is a new_market event
				// Note: Polymarket wraps in array
				var events []json.RawMessage
				if err := json.Unmarshal(msg, &events); err == nil {
					for _, rawEvent := range events {
						var newMarket NewMarketEvent
						if err := json.Unmarshal(rawEvent, &newMarket); err == nil && newMarket.EventType == "new_market" {
							log.Printf("Discovered new market: %s, tokens: %v", newMarket.ConditionID, newMarket.ClobTokenIDs)
							completenessArb.RegisterMarket(newMarket.ConditionID, newMarket.ClobTokenIDs)
							// Subscribe to updates for these new tokens
							_ = client.Subscribe(newMarket.ClobTokenIDs)
						}
					}
				}

				// 3. Feed into CLOB Manager
				if err := clobManager.ProcessWebSocketMessage(msg); err != nil {
					log.Printf("Failed to process message into CLOB: %v", err)
				}
			}
		}
	}()

	// Handle graceful shutdown on OS signal
	go func() {
		sig := <-sigs
		log.Printf("Received signal: %v, shutting down...", sig)
		cancel()
	}()

	log.Println("Engine is running. Listening for live orderbook updates...")

	<-ctx.Done()
	log.Println("Shutdown complete.")
}
