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

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
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

// PingAuth checks if the configured API credentials are valid
func (c *Client) PingAuth() error {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	method := http.MethodGet
	requestPath := "/auth" // Polymarket CLOB has an /auth endpoint for checking credentials
	message := timestamp + method + requestPath

	decodedSecret, err := base64.StdEncoding.DecodeString(c.apiSecret)
	if err != nil {
		return fmt.Errorf("failed to decode api secret: %w", err)
	}

	h := hmac.New(sha256.New, decodedSecret)
	h.Write([]byte(message))
	apiSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	reqURL := fmt.Sprintf("%s%s", c.endpoint, requestPath)
	req, err := http.NewRequest(method, reqURL, nil)
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

// GenerateKeysFromWallet automatically derives Polymarket API credentials using an EIP-712 ClobAuth signature
func GenerateKeysFromWallet(address, privKeyHex string) (apiKey, secret, passphrase string, err error) {
	// Remove 0x prefix if present
	if len(privKeyHex) > 2 && privKeyHex[:2] == "0x" {
		privKeyHex = privKeyHex[2:]
	}

	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid private key: %w", err)
	}

	timestampStr := fmt.Sprintf("%d", time.Now().Unix())
	nonce := fmt.Sprintf("%d", time.Now().UnixMilli())

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"ClobAuth": []apitypes.Type{
				{Name: "address", Type: "address"},
				{Name: "timestamp", Type: "string"},
				{Name: "nonce", Type: "uint256"},
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "ClobAuth",
		Domain: apitypes.TypedDataDomain{
			Name:    "ClobAuthDomain",
			Version: "1",
			ChainId: math.NewHexOrDecimal256(137), // Polygon Mainnet
		},
		Message: apitypes.TypedDataMessage{
			"address":   address,
			"timestamp": timestampStr,
			"nonce":     nonce,
			"message":   "This message attests that I control the given wallet",
		},
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", "", "", fmt.Errorf("failed to hash domain: %w", err)
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to hash message: %w", err)
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	hash := crypto.Keccak256Hash(rawData)

	sig, err := crypto.Sign(hash.Bytes(), privKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to sign: %w", err)
	}

	sig[64] += 27 // Adjust V value for Ethereum
	signature := fmt.Sprintf("0x%x", sig)

	req, err := http.NewRequest("POST", defaultCLOBEndpoint+"/auth/api-key", nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("POLY_ADDRESS", address)
	req.Header.Set("POLY_SIGNATURE", signature)
	req.Header.Set("POLY_TIMESTAMP", timestampStr)
	req.Header.Set("POLY_NONCE", nonce)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("API rejected the signature. Status: %d", resp.StatusCode)
	}

	var result struct {
		ApiKey     string `json:"apiKey"`
		Secret     string `json:"secret"`
		Passphrase string `json:"passphrase"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", fmt.Errorf("failed to parse Polymarket response: %w", err)
	}

	return result.ApiKey, result.Secret, result.Passphrase, nil
}
