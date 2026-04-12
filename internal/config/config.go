package config

import (
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

type Config struct {
	PolygonPrivateKey  string
	PublicAddress      common.Address
	PolyApiKey         string
	PolyApiSecret      string
	PolyApiPassphrase  string
	RelayerAPIKey      string
	RelayerAPIKeyAddr  string
	MinNetProfitMargin float64
	TelegramBotToken   string
	TelegramChatID     string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	privKey := os.Getenv("POLYGON_PRIVATE_KEY")
	pubAddrStr := os.Getenv("PUBLIC_ADDRESS")
	polyApiKey := os.Getenv("POLY_API_KEY")
	polyApiSecret := os.Getenv("POLY_API_SECRET")
	polyApiPassphrase := os.Getenv("POLY_API_PASSPHRASE")

	relayerApiKey := os.Getenv("RELAYER_API_KEY")
	relayerApiKeyAddr := os.Getenv("RELAYER_API_KEY_ADDRESS")
	if relayerApiKeyAddr == "" {
		relayerApiKeyAddr = "0xC6dcCCB919Bf7EEb49864152832F0D8E6203199F"
	}

	var pubAddr common.Address
	if pubAddrStr != "" {
		pubAddr = common.HexToAddress(pubAddrStr)
	}

	minNetProfitMarginStr := os.Getenv("MIN_NET_PROFIT_MARGIN")
	minNetProfitMargin := 0.03
	if minNetProfitMarginStr != "" {
		if parsed, err := strconv.ParseFloat(minNetProfitMarginStr, 64); err == nil {
			minNetProfitMargin = parsed
		}
	}

	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	return &Config{
		PolygonPrivateKey:  privKey,
		PublicAddress:      pubAddr,
		PolyApiKey:         polyApiKey,
		PolyApiSecret:      polyApiSecret,
		PolyApiPassphrase:  polyApiPassphrase,
		RelayerAPIKey:      relayerApiKey,
		RelayerAPIKeyAddr:  relayerApiKeyAddr,
		MinNetProfitMargin: minNetProfitMargin,
		TelegramBotToken:   telegramBotToken,
		TelegramChatID:     telegramChatID,
	}
}
