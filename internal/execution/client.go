package execution

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultCLOBEndpoint = "https://clob.polymarket.com"

// Client is a REST API client for the Polymarket CLOB
type Client struct {
	endpoint      string
	httpClient    *http.Client
	apiKey        string
	apiSecret     string
	apiPassphrase string
}

// NewClient creates a new CLOB API client
func NewClient(apiKey, apiSecret, apiPassphrase string) *Client {
	return &Client{
		endpoint: defaultCLOBEndpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		apiPassphrase: apiPassphrase,
	}
}

// OrderPayload represents the JSON payload expected by the /order endpoint
type OrderPayload struct {
	Order     *Order `json:"order"`
	Signature string `json:"signature"`
}

// PlaceOrder sends a signed order to the CLOB /order endpoint
func (c *Client) PlaceOrder(order *Order, signature string) error {
	payload := OrderPayload{
		Order:     order,
		Signature: signature,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal order payload: %w", err)
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	method := http.MethodPost
	requestPath := "/order"
	message := timestamp + method + requestPath + string(data)

	decodedSecret, err := base64.StdEncoding.DecodeString(c.apiSecret)
	if err != nil {
		return fmt.Errorf("failed to decode api secret: %w", err)
	}

	h := hmac.New(sha256.New, decodedSecret)
	h.Write([]byte(message))
	apiSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	reqURL := fmt.Sprintf("%s%s", c.endpoint, requestPath)
	req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("POLY-API-KEY", c.apiKey)
	req.Header.Set("POLY-API-SIGNATURE", apiSignature)
	req.Header.Set("POLY-API-TIMESTAMP", timestamp)
	req.Header.Set("POLY-API-PASSPHRASE", c.apiPassphrase)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}
