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
	maxBackoff        = 30 * time.Second
	initialBackoff    = 1 * time.Second
)

// WsClient handles the connection to the Polymarket WebSocket.
type WsClient struct {
	url       string
	conn      *websocket.Conn
	marketIDs []string // Store for resubscription
}

// NewWsClient creates a new WebSocket client.
func NewWsClient(endpoint string) *WsClient {
	if endpoint == "" {
		endpoint = DefaultWSEndpoint
	}
	return &WsClient{
		url:       endpoint,
		marketIDs: []string{},
	}
}

// Connect establishes the WebSocket connection with exponential backoff.
func (c *WsClient) Connect(ctx context.Context) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return err
	}

	backoff := initialBackoff

	for {
		log.Printf("Connecting to Polymarket WebSocket at %s", u.String())
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
		if err == nil {
			c.conn = conn
			log.Println("WebSocket connected successfully")
			backoff = initialBackoff // reset on success

			// Resubscribe if we have active subscriptions
			if len(c.marketIDs) > 0 {
				if subErr := c.Subscribe(c.marketIDs); subErr != nil {
					log.Printf("Resubscription failed: %v", subErr)
				}
			}

			// Start PING loop
			go c.pingLoop(ctx)
			return nil
		}

		log.Printf("WebSocket connection failed: %v. Retrying in %v...", err, backoff)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (c *WsClient) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if c.conn != nil {
				if err := c.conn.WriteMessage(websocket.TextMessage, []byte("PING")); err != nil {
					log.Printf("Failed to send PING: %v", err)
					return // exit ping loop on error, connection will drop
				}
			}
		}
	}
}

// Listen starts a blocking loop that reads messages from the WebSocket.
// It automatically handles reconnections if the connection drops.
func (c *WsClient) Listen(ctx context.Context, msgChan chan<- []byte) {
	for {
		if c.conn != nil {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v. Attempting to reconnect...", err)
				c.conn.Close()
				c.conn = nil
				if err := c.Connect(ctx); err != nil {
					log.Printf("Failed to reconnect: %v", err)
					return // Give up if context cancelled or fatal error
				}
				continue
			}
			log.Printf("WS RECV: %s", string(message))
			msgChan <- message
		}

		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping listener.")
			if c.conn != nil {
				c.conn.Close()
			}
			return
		default:
			if c.conn == nil {
				time.Sleep(100 * time.Millisecond) // avoid tight loop if reconnect fails
			}
		}
	}
}

// Subscribe sends a subscription message for specific market IDs.
func (c *WsClient) Subscribe(marketIDs []string) error {
	c.marketIDs = marketIDs // save for resubscription

	// Pre-warm the metadata cache for these assets
	LoadMetadataForAssets(marketIDs)

	subMsg := map[string]interface{}{
		"assets_ids":             marketIDs,
		"type":                   "market",
		"custom_feature_enabled": true,
	}

	payload, err := json.Marshal(subMsg)
	if err != nil {
		return err
	}

	log.Printf("Sending subscription: %s", string(payload))
	if c.conn != nil {
		return c.conn.WriteMessage(websocket.TextMessage, payload)
	}
	return nil
}
