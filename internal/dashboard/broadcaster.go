package dashboard

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"allele/internal/abi"
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
}

func NewBroadcaster(pm *plugin.Manager) *Broadcaster {
	return &Broadcaster{
		conns: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			// In production, we should restrict CheckOrigin to the UI's host.
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		pm: pm,
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

			// Save to DB
			if err := storage.SetPluginConfig(req.PluginName, req.Key, req.Value); err != nil {
				b.sendError(conn, "update_config_error", err.Error())
				continue
			}

			// Broadcast plugin_status event so all panels know a config changed
			b.Broadcast("plugin_status", map[string]string{
				"plugin": req.PluginName,
				"status": "config_updated",
			})

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

	for _, m := range manifests {
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
