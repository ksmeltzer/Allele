package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	privKeyHex := os.Getenv("POLYGON_PRIVATE_KEY")
	address := os.Getenv("PUBLIC_ADDRESS")

	if privKeyHex == "" || address == "" {
		log.Fatal("POLYGON_PRIVATE_KEY and PUBLIC_ADDRESS must be set in .env")
	}

	// Remove 0x prefix if present
	privKeyHex = strings.TrimPrefix(privKeyHex, "0x")

	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		log.Fatalf("Invalid private key: %v", err)
	}

	timestampStr := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "0"

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
			ChainId: math.NewHexOrDecimal256(137),
		},
		Message: apitypes.TypedDataMessage{
			"address":   address,
			"timestamp": timestampStr,
			"nonce":     nonce,
			"message":   "This message attests that I control the given wallet",
		},
	}

	// Using HashStruct handles EIP712 hash generation.
	// We want to combine the domain separator and message hash.
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		log.Fatalf("Failed to hash domain: %v", err)
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		log.Fatalf("Failed to hash message: %v", err)
	}

	// EIP-712 standard: "\x19\x01" + domainSeparator + typedDataHash
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	hash := crypto.Keccak256Hash(rawData)

	sig, err := crypto.Sign(hash.Bytes(), privKey)
	if err != nil {
		log.Fatalf("Failed to sign: %v", err)
	}

	// Add 27 to V
	sig[64] += 27
	signature := fmt.Sprintf("0x%x", sig)

	// Make request
	req, err := http.NewRequest("POST", "https://clob.polymarket.com/auth/api-key", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("POLY_ADDRESS", address)
	req.Header.Set("POLY_SIGNATURE", signature)
	req.Header.Set("POLY_TIMESTAMP", timestampStr)
	req.Header.Set("POLY_NONCE", nonce)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ApiKey     string `json:"apiKey"`
		Secret     string `json:"secret"`
		Passphrase string `json:"passphrase"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v. Body: %s", err, string(body))
	}

	// Automate updating the .env file
	envMap, err := godotenv.Read(".env")
	if err != nil {
		log.Printf("Warning: Could not read .env file for writing, will just print to console: %v", err)
	} else {
		envMap["POLY_API_KEY"] = result.ApiKey
		envMap["POLY_API_SECRET"] = result.Secret
		envMap["POLY_API_PASSPHRASE"] = result.Passphrase

		if err := godotenv.Write(envMap, ".env"); err != nil {
			log.Printf("Warning: Failed to write to .env file: %v", err)
		} else {
			fmt.Println("✅ Successfully injected API credentials into .env file!")
		}
	}

	fmt.Println("======================================")
	fmt.Println("🎉 Polymarket API Key Generated! 🎉")
	fmt.Println("======================================")
	fmt.Printf("API_KEY    : %s\n", result.ApiKey)
	fmt.Printf("SECRET     : %s\n", result.Secret)
	fmt.Printf("PASSPHRASE : %s\n", result.Passphrase)
	fmt.Println("======================================")
}
