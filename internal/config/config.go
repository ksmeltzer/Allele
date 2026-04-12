package config

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

type Config struct {
	PolygonPrivateKey string
	PublicAddress     common.Address
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	privKey := os.Getenv("POLYGON_PRIVATE_KEY")
	pubAddrStr := os.Getenv("PUBLIC_ADDRESS")

	var pubAddr common.Address
	if pubAddrStr != "" {
		pubAddr = common.HexToAddress(pubAddrStr)
	}

	return &Config{
		PolygonPrivateKey: privKey,
		PublicAddress:     pubAddr,
	}
}
