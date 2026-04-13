package dashboard

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"allele/internal/abi"
	"allele/internal/storage"

	"github.com/gorilla/websocket"
)

type Broadcaster struct {
	conns    map[*websocket.Conn]bool
	mutex    sync.Mutex
	upgrader websocket.Upgrader
	// In a real implementation this would query the loaded WASM modules.
	// For now, we mock the manifest data for the UI to consume.
	mockManifests []abi.Manifest
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		conns: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		mockManifests: []abi.Manifest{
			{
				Name:        "allele-exchange-polymarket",
				Version:     "v1.0.0",
				Description: "Polymarket Exchange Adapter",
				Author:      "Allele Org",
				Dependencies: []abi.Dependency{},
				Config: []abi.ConfigField{
					{Key: "POLY_API_KEY", Type: "secret", Description: "Polymarket API Key", Required: true},
					{Key: "POLY_API_SECRET", Type: "secret", Description: "Polymarket API Secret", Required: true},
					{Key: "POLY_API_PASSPHRASE", Type: "secret", Description: "Polymarket API Passphrase", Required: true},
				},
			},
			{
				Name:        "allele-strategy-cross-market",
				Version:     "v1.0.0",
				Description: "Cross-Market Correlation Arbitrage",
				Author:      "Allele Org",
				Dependencies: []abi.Dependency{
					{Name: "allele-exchange-polymarket", Type: "exchange", Version: ">=v1.0.0"},
				},
				Config: []abi.ConfigField{
					{Key: "MIN_SPREAD", Type: "string", Description: "Minimum spread to execute", Required: true},
					{Key: "EXPERIMENTAL_MODE", Type: "boolean", Description: "Enable risky trades", Required: false},
				},
			},
		},
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (b *Broadcaster) Start(port string) {
	expectedToken, _ := storage.GetPluginConfig("system", "BROADCASTER_AUTH_TOKEN")
	if expectedToken == "" {
		expectedToken = "dev-token" // Fallback for dev mode
		storage.SetPluginConfig("system", "BROADCASTER_AUTH_TOKEN", expectedToken)
		log.Printf("Warning: BROADCASTER_AUTH_TOKEN not set. Defaulting to 'dev-token'")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("auth_token")
		if token == "" || token != expectedToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := b.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		b.mutex.Lock()
		b.conns[conn] = true
		b.mutex.Unlock()

		go func() {
			defer func() {
				b.mutex.Lock()
				delete(b.conns, conn)
				b.mutex.Unlock()
				conn.Close()
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	})

	mux.HandleFunc("/api/plugins", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Attach current values from DB to the mock response
		type ConfigFieldWithValue struct {
			abi.ConfigField
			Value string `json:"value"`
		}
		type ManifestWithValues struct {
			abi.Manifest
			Config []ConfigFieldWithValue `json:"config"`
		}

		var resp []ManifestWithValues
		for _, m := range b.mockManifests {
			var newConfig []ConfigFieldWithValue
			for _, c := range m.Config {
				val, _ := storage.GetPluginConfig(m.Name, c.Key)
				// Mask secret values when sending to frontend
				if c.Type == "secret" && val != "" {
					val = "********"
				}
				newConfig = append(newConfig, ConfigFieldWithValue{
					ConfigField: c,
					Value:       val,
				})
			}
			resp = append(resp, ManifestWithValues{
				Manifest: m,
				Config:   newConfig,
			})
		}

		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/plugins/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			PluginName string `json:"plugin_name"`
			Key        string `json:"key"`
			Value      string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// If the user submits masked value, ignore it
		if req.Value == "********" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if err := storage.SetPluginConfig(req.PluginName, req.Key, req.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Starting Broadcaster (API & WS) on localhost%s", port)
	if err := http.ListenAndServe("127.0.0.1"+port, enableCORS(mux)); err != nil {
		log.Fatalf("Broadcaster server failed: %v", err)
	}
}

func (b *Broadcaster) Broadcast(payload interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for conn := range b.conns {
		if err := conn.WriteJSON(payload); err != nil {
			log.Printf("Broadcast error: %v", err)
			conn.Close()
			delete(b.conns, conn)
		}
	}
}
