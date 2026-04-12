package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

type RelayerClient struct {
	apiKey        string
	apiKeyAddress string
	httpClient    *http.Client
}

func NewRelayerClient(apiKey, apiKeyAddress string) *RelayerClient {
	return &RelayerClient{
		apiKey:        apiKey,
		apiKeyAddress: apiKeyAddress,
		httpClient:    &http.Client{},
	}
}

type MergeRequest struct {
	ConditionID string `json:"conditionId"`
	Amount      string `json:"amount"`
}

func (c *RelayerClient) Merge(conditionID string, amount *big.Int) error {
	url := "https://relayer.polymarket.com/merge" // default polymarket relayer merge endpoint

	reqBody := MergeRequest{
		ConditionID: conditionID,
		Amount:      amount.String(),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal merge request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("RELAYER_API_KEY", c.apiKey)
	req.Header.Set("RELAYER_API_KEY_ADDRESS", c.apiKeyAddress)
	req.Header.Set("Signer Address", c.apiKeyAddress)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("relayer API returned non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}
