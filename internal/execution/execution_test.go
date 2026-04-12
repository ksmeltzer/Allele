package execution

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestEIP712Signing(t *testing.T) {
	// Generate dummy private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create dummy order
	order := Order{
		Salt:          big.NewInt(123456789),
		Maker:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Signer:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
		TokenId:       big.NewInt(10),
		MakerAmount:   big.NewInt(100),
		TakerAmount:   big.NewInt(200),
		Expiration:    big.NewInt(9999999999),
		Nonce:         big.NewInt(1),
		FeeRateBps:    big.NewInt(0),
		Side:          0, // Buy
		SignatureType: 1, // EOA
	}

	// Sign order
	sig, err := SignOrder(order, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign order: %v", err)
	}

	// Ensure signature is exactly 65 bytes long (R, S, V)
	if len(sig) != 65 {
		t.Errorf("Expected signature to be 65 bytes long, got %d", len(sig))
	}
}

func TestJSONMarshalling(t *testing.T) {
	order := Order{
		Salt:          big.NewInt(123456789),
		Maker:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Signer:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Taker:         common.HexToAddress("0x0000000000000000000000000000000000000000"),
		TokenId:       big.NewInt(10),
		MakerAmount:   big.NewInt(100),
		TakerAmount:   big.NewInt(200),
		Expiration:    big.NewInt(9999999999),
		Nonce:         big.NewInt(1),
		FeeRateBps:    big.NewInt(0),
		Side:          0,
		SignatureType: 1,
	}

	payload := OrderPayload{
		Order:     &order,
		Signature: "0xdummy_signature_string",
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal OrderPayload: %v", err)
	}

	t.Logf("Marshalled JSON Payload:\n%s", string(data))

	// Some basic verifications that the strings appear correctly
	jsonStr := string(data)
	expectedStrings := []string{
		`"salt": "123456789"`,
		`"maker": "0x1234567890123456789012345678901234567890"`,
		`"tokenId": "10"`,
		`"makerAmount": "100"`,
		`"signature": "0xdummy_signature_string"`,
	}

	for _, expected := range expectedStrings {
		if !contains(jsonStr, expected) {
			t.Errorf("Expected JSON to contain %s", expected)
		}
	}
}

func contains(s, substr string) bool {
	// A simple helper for testing since strings.Contains requires an import
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
