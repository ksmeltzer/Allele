package polymarket

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Default Polymarket CLOB WebSocket endpoint
	DefaultWSEndpoint = "wss://ws-subscriptions-clob.polymarket.com/ws/market"
)

// WsClient handles the connection to the Polymarket WebSocket.
type WsClient struct {
	url  string
	conn *websocket.Conn
}

// NewWsClient creates a new WebSocket client.
func NewWsClient(endpoint string) *WsClient {
	if endpoint == "" {
		endpoint = DefaultWSEndpoint
	}
	return &WsClient{
		url: endpoint,
	}
}

// Connect establishes the WebSocket connection.
func (c *WsClient) Connect(ctx context.Context) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return err
	}

	log.Printf("Connecting to Polymarket WebSocket at %s", u.String())
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return err
	}

	c.conn = conn

	// Start PING loop (required every 10 seconds per Polymarket docs)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.conn.WriteMessage(websocket.TextMessage, []byte("PING")); err != nil {
					log.Printf("Failed to send PING: %v", err)
					return
				}
			}
		}
	}()

	return nil
}

// Listen starts a blocking loop that reads messages from the WebSocket.
func (c *WsClient) Listen(ctx context.Context, msgChan chan<- []byte) {
	defer c.conn.Close()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping listener.")
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}
			msgChan <- message
		}
	}
}

// Subscribe sends a subscription message for specific market IDs.
func (c *WsClient) Subscribe(marketIDs []string) error {
	subMsg := map[string]interface{}{
		"assets_ids": marketIDs,
		"type":       "market",
		"custom_feature_enabled": true,
	}

	payload, err := json.Marshal(subMsg)
	if err != nil {
		return err
	}

	log.Printf("Sending subscription: %s", string(payload))
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}
