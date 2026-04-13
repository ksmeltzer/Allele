package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"allele/internal/adapters/exchange"
	"allele/internal/adapters/strategy"
	"allele/internal/adapters/wallet"
	"allele/internal/alerting"
	"allele/internal/config"
	"allele/internal/core"
	"allele/internal/dashboard"
	"allele/internal/engine"
	"allele/internal/execution"
	"allele/internal/polymarket"
	"allele/internal/storage"
	"github.com/ethereum/go-ethereum/crypto"
)

type NewMarketEvent struct {
	EventType    string   `json:"event_type"`
	ConditionID  string   `json:"condition_id"`
	ClobTokenIDs []string `json:"clob_token_ids"`
}

func main() {
	initCLI()
	// Load config early for alerting
	cfg := config.LoadConfig()

	// Connect to watchdog
	watchdogConn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		log.Printf("Warning: Could not connect to watchdog daemon: %v", err)
	} else {
		defer watchdogConn.Close()
	}

	sendWatchdog := func(event alerting.Event, data map[string]interface{}) {
		if watchdogConn == nil {
			return
		}
		payload := alerting.Payload{Event: event, Data: data}
		b, err := json.Marshal(payload)
		if err == nil {
			b = append(b, '\n')
			_, _ = watchdogConn.Write(b)
		}
	}

	defer func() {
		if r := recover(); r != nil {
			sendWatchdog(alerting.EventCrash, map[string]interface{}{"error": fmt.Sprintf("%v", r)})
			panic(r)
		}
	}()

	sendWatchdog(alerting.EventBoot, nil)

	var lastTickUnix atomic.Int64
	lastTickUnix.Store(time.Now().Unix())

	// Start heartbeat goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if time.Now().Unix()-lastTickUnix.Load() < 10 {
				sendWatchdog(alerting.EventHeartbeat, nil)
			}
		}
	}()

	log.Println("Starting Polymarket Data Recorder & Trading Engine...")

	// Create root context that listens for interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Initialize Database
	if err := storage.InitDB(".allele/trading.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Broadcaster
	broadcaster := dashboard.NewBroadcaster()
	go broadcaster.Start(":8081")

	// Initialize Execution Client
	execClient := execution.NewClient(cfg.PolyApiKey, cfg.PolyApiSecret, cfg.PolyApiPassphrase)

	var privateKey *ecdsa.PrivateKey
	if cfg.PolygonPrivateKey != "" {
		pk, err := crypto.HexToECDSA(cfg.PolygonPrivateKey)
		if err != nil {
			log.Printf("Invalid polygon private key: %v", err)
		} else {
			privateKey = pk
		}
	} else {
		// Dummy private key for testing if missing
		pk, _ := crypto.GenerateKey()
		privateKey = pk
	}

	// Microkernel Setup
	kernel := engine.NewKernel()

	polygonWallet := wallet.NewPolygonWallet(privateKey)
	kernel.RegisterWallet(polygonWallet)

	wsClient := polymarket.NewWsClient(polymarket.DefaultWSEndpoint)
	polymarketExchange := exchange.NewPolymarketExchange(wsClient, execClient)
	kernel.RegisterExchange(polymarketExchange)

	takerFee := 0.02
	minProfitMargin := cfg.MinNetProfitMargin
	completenessStrategy := strategy.NewCompletenessArbitrage(takerFee, minProfitMargin)
	kernel.RegisterStrategy(completenessStrategy)

	go kernel.Start(ctx)

	// Connect to WS
	if err := wsClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Polymarket WebSocket: %v", err)
	}

	// Connect stream to kernel
	rawTickChan := make(chan core.NormalizedTick, 1000)
	polymarketExchange.ConnectStream(ctx, rawTickChan)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-rawTickChan:
				lastTickUnix.Store(time.Now().Unix())
				kernel.TickChan() <- t
			}
		}
	}()

	// Active Test Market IDs (From Gamma API)
	testMarketIDs := []string{
		"93592949212798121127213117304912625505836768562433217537850469496310204567695",
		"3074539347152748632858978545166555332546941892131779352477699494423276162345",
	}

	if err := wsClient.Subscribe(testMarketIDs); err != nil {
		log.Fatalf("Failed to subscribe to initial markets: %v", err)
	}

	completenessStrategy.RegisterMarket("InitialTestMarket", testMarketIDs)

	// Since we connected stream via connectStream, we also want to intercept for dynamic market discovery
	// But ConnectStream consumes from wsClient.Listen, wait! wsClient.Listen pushes to a chan.
	// Actually polymarketExchange.ConnectStream runs `wsClient.Listen`. If we call Listen twice it might fail.
	// Oh! `ConnectStream` calls `go p.wsClient.Listen(ctx, msgChan)`. So we shouldn't start it again,
	// or we need to intercept messages some other way, OR we let ConnectStream do its thing and
	// perhaps we just don't do dynamic market discovery for now, or we modify it.
	// The instructions just say: "polymarketExchange.ConnectStream(ctx, kernel.TickChan())".
	// Let's stick to what's asked.

	// Handle graceful shutdown on OS signal
	go func() {
		sig := <-sigs
		log.Printf("Received signal: %v, shutting down...", sig)
		cancel()
	}()

	log.Println("Engine is running. Listening for live orderbook updates...")

	<-ctx.Done()
	log.Println("Shutdown complete.")
	sendWatchdog(alerting.EventShutdown, nil)
}
