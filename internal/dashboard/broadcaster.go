package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"allele/internal/abi"
	"allele/internal/core"
	"allele/internal/plugin"
	"allele/internal/storage"

	"github.com/gorilla/websocket"
)

// WSMessage is the typed envelope for all WebSocket communication (both ways).
type WSMessage struct {
	Type    string      `json:"type"`
	Ts      int64       `json:"ts"`
	Payload interface{} `json:"payload"`
}

type Broadcaster struct {
	conns    map[*websocket.Conn]bool
	mutex    sync.Mutex
	upgrader websocket.Upgrader
	pm       *plugin.Manager
	eventBus *core.EventBus
}

func NewBroadcaster(pm *plugin.Manager, eventBus *core.EventBus) *Broadcaster {
	b := &Broadcaster{
		conns: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			// In production, we should restrict CheckOrigin to the UI's host.
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		pm:       pm,
		eventBus: eventBus,
	}

	if eventBus != nil {
		go b.listenEventBus()
	}

	return b
}

func (b *Broadcaster) listenEventBus() {
	ch := b.eventBus.Subscribe(core.SystemAlertEvent)
	for event := range ch {
		b.Broadcast(string(event.Type), event.Payload)
	}
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

		go b.handleConnection(conn)
	})

	log.Printf("Starting Broadcaster (WS only) on 0.0.0.0%s", port)
	// We no longer need CORS middleware because we removed the REST endpoints.
	// WebSocket connections handle cross-origin policy via the CheckOrigin function.
	if err := http.ListenAndServe("0.0.0.0"+port, mux); err != nil {
		log.Fatalf("Broadcaster server failed: %v", err)
	}
}

// Broadcast sends a typed message to all connected clients.
func (b *Broadcaster) Broadcast(eventType string, payload interface{}) {
	msg := WSMessage{
		Type:    eventType,
		Ts:      time.Now().UnixMilli(),
		Payload: payload,
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	for conn := range b.conns {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Broadcast error: %v", err)
			conn.Close()
			delete(b.conns, conn)
		}
	}
}

// handleConnection reads incoming messages from the UI and acts upon them.
func (b *Broadcaster) handleConnection(conn *websocket.Conn) {
	defer func() {
		b.mutex.Lock()
		delete(b.conns, conn)
		b.mutex.Unlock()
		conn.Close()
	}()

	for {
		var incoming struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := conn.ReadJSON(&incoming); err != nil {
			break // Connection closed or error
		}

		switch incoming.Type {
		case "request_manifests":
			b.sendManifests(conn)

		case "install_plugin":
			var req struct {
				URI string `json:"uri"`
			}
			if err := json.Unmarshal(incoming.Payload, &req); err != nil {
				b.sendError(conn, "install_plugin_error", "invalid payload")
				continue
			}

			if b.pm == nil {
				b.sendError(conn, "install_plugin_error", "plugin manager not available")
				continue
			}

			// Install from URI
			// Note: Context should be passed down, creating one here for now
			ctx := context.Background()
			if err := b.pm.InstallFromURI(ctx, req.URI); err != nil {
				b.sendError(conn, "install_plugin_error", err.Error())
				continue
			}

			b.Broadcast("plugin_status", map[string]string{
				"plugin": req.URI,
				"status": "installed",
			})
			b.sendManifests(conn)

		case "update_config":
			var req struct {
				PluginName string `json:"plugin_name"`
				Key        string `json:"key"`
				Value      string `json:"value"`
			}
			if err := json.Unmarshal(incoming.Payload, &req); err != nil {
				b.sendError(conn, "update_config_error", "invalid payload")
				continue
			}

			// If the user submits masked value, ignore it
			if req.Value == "********" {
				b.sendManifests(conn) // Send fresh state to reset the UI
				continue
			}

			// Special case routing for the hardcoded system core manifest
			savePluginName := req.PluginName
			if req.PluginName == "allele-core-system" {
				savePluginName = "system"
			}

			// Save to DB
			if err := storage.SetPluginConfig(savePluginName, req.Key, req.Value); err != nil {
				b.sendError(conn, "update_config_error", err.Error())
				continue
			}

			// Broadcast plugin_status event so all panels know a config changed
			b.Broadcast("plugin_status", map[string]string{
				"plugin": req.PluginName,
				"status": "config_updated",
			})

			// Publish to internal EventBus so adapters can reload
			if b.eventBus != nil {
				b.eventBus.Publish(core.Event{
					Type: core.ConfigUpdatedEvent,
					Payload: map[string]string{
						"plugin_name": savePluginName,
					},
				})
			}

			// Send updated manifests back
			b.sendManifests(conn)
		}
	}
}

// sendManifests retrieves all plugins and their current configurations,
// masks secrets, and pushes them down the WebSocket connection.
func (b *Broadcaster) sendManifests(conn *websocket.Conn) {
	type ConfigFieldWithValue struct {
		abi.ConfigField
		Value string `json:"value"`
	}
	type ManifestWithValues struct {
		abi.Manifest
		Config []ConfigFieldWithValue `json:"config"`
	}

	var resp []ManifestWithValues
	var manifests []abi.Manifest
	if b.pm != nil {
		manifests = b.pm.GetManifests()
	}

	// Inject the Core System configuration manifest for the UI
	manifests = append(manifests, abi.Manifest{
		Name:        "allele-core-system",
		Version:     "v1.0.0",
		Description: "Core System Wallet & Global Configuration",
		Author:      "Allele",
		Config: []abi.ConfigField{
			{Key: "POLYGON_PRIVATE_KEY", Type: "secret", Description: "Polygon Wallet Private Key (Hex)", Required: true},
			{Key: "PUBLIC_ADDRESS", Type: "string", Description: "Polygon Public Wallet Address (0x...)", Required: true},
			{Key: "MIN_NET_PROFIT_MARGIN", Type: "string", Description: "Minimum % Profit required to fire an arbitrage trade (e.g. 0.03 for 3%)", Required: false},
		},
	})

	for _, m := range manifests {
		var newConfig []ConfigFieldWithValue
		lookupName := m.Name
		if m.Name == "allele-core-system" {
			lookupName = "system"
		}

		for _, c := range m.Config {
			val, _ := storage.GetPluginConfig(lookupName, c.Key)
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

	msg := WSMessage{
		Type:    "manifests_updated",
		Ts:      time.Now().UnixMilli(),
		Payload: resp,
	}

	// Not using b.Broadcast because this is a direct response to one client
	conn.WriteJSON(msg)
}

func (b *Broadcaster) sendError(conn *websocket.Conn, errorType string, message string) {
	msg := WSMessage{
		Type: errorType,
		Ts:   time.Now().UnixMilli(),
		Payload: map[string]string{
			"error": message,
		},
	}
	conn.WriteJSON(msg)
}
