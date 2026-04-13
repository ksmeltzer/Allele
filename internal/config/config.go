package config

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"allele/internal/storage"
)

// Config represents the system-wide configuration loaded dynamically from SQLite and the Secure Vault.
type Config struct {
	PolygonPrivateKey  string
	PublicAddress      common.Address
	PolyApiKey         string
	PolyApiSecret      string
	PolyApiPassphrase  string
	RelayerAPIKey      string
	RelayerAPIKeyAddr  string
	MinNetProfitMargin float64
}

// LoadConfig reads the system parameters directly from the SQLite database.
func LoadConfig() *Config {
	// 1. Fetch non-secret parameters from the generic SQLite plugin config table
	pubAddrStr, _ := storage.GetPluginConfig("system", "PUBLIC_ADDRESS")
	relayerApiKeyAddr, _ := storage.GetPluginConfig("system", "RELAYER_API_KEY_ADDRESS")
	if relayerApiKeyAddr == "" {
		relayerApiKeyAddr = "0xC6dcCCB919Bf7EEb49864152832F0D8E6203199F"
	}

	var pubAddr common.Address
	if pubAddrStr != "" {
		pubAddr = common.HexToAddress(pubAddrStr)
	}

	minNetProfitMarginStr, _ := storage.GetPluginConfig("system", "MIN_NET_PROFIT_MARGIN")
	minNetProfitMargin := 0.03
	if minNetProfitMarginStr != "" {
		if parsed, err := strconv.ParseFloat(minNetProfitMarginStr, 64); err == nil {
			minNetProfitMargin = parsed
		}
	}

	// 2. Fetch Secrets from the AES-GCM Encrypted Vault (if available)
	// (Note: To fetch from Vault we need the vault instance, but for now we
	//  will just initialize empty string for secrets here; the Engine/Adapters
	//  will fetch directly from the Vault during execution, rather than global config).
	var privKey, polyApiKey, polyApiSecret, polyApiPassphrase, relayerApiKey string

	return &Config{
		PolygonPrivateKey:  privKey,
		PublicAddress:      pubAddr,
		PolyApiKey:         polyApiKey,
		PolyApiSecret:      polyApiSecret,
		PolyApiPassphrase:  polyApiPassphrase,
		RelayerAPIKey:      relayerApiKey,
		RelayerAPIKeyAddr:  relayerApiKeyAddr,
		MinNetProfitMargin: minNetProfitMargin,
	}
}
