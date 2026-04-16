package polymarket

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GammaMarket struct {
	Question     string `json:"question"`
	ClobTokenIds string `json:"clobTokenIds"` // JSON string array
	Outcomes     string `json:"outcomes"`     // JSON string array
}

var (
	metadataCache = make(map[string]string)
)

// LoadMetadataForAssets asynchronously fetches metadata for multiple assets to populate the cache
func LoadMetadataForAssets(assetIDs []string) {
	for _, id := range assetIDs {
		if _, ok := metadataCache[id]; !ok {
			go GetAssetMetadata(id) // Fire and forget, will cache it for future use
		}
	}
}

// GetAssetMetadata fetches the human-readable name for a Polymarket Asset ID.
// It returns a cached string if available, otherwise queries the Gamma API.
func GetAssetMetadata(assetID string) string {
	if name, ok := metadataCache[assetID]; ok {
		return name
	}

	url := fmt.Sprintf("https://gamma-api.polymarket.com/markets?clob_token_ids=%s", assetID)
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return assetID
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return assetID
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return assetID
	}

	var markets []GammaMarket
	if err := json.Unmarshal(b, &markets); err != nil || len(markets) == 0 {
		return assetID
	}

	market := markets[0]

	var clobIDs []string
	if err := json.Unmarshal([]byte(market.ClobTokenIds), &clobIDs); err != nil {
		return assetID
	}

	var outcomes []string
	if err := json.Unmarshal([]byte(market.Outcomes), &outcomes); err != nil {
		return assetID
	}

	// Cache all outcomes for this market
	for i, id := range clobIDs {
		if i < len(outcomes) {
			metadataCache[id] = fmt.Sprintf("%s (%s)", market.Question, outcomes[i])
		} else {
			metadataCache[id] = market.Question
		}
	}

	if name, ok := metadataCache[assetID]; ok {
		return name
	}

	return assetID
}
