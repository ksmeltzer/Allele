package execution

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

var (
	PolymarketDomain = apitypes.TypedDataDomain{
		Name:              "Polymarket Exchange",
		Version:           "1",
		ChainId:           math.NewHexOrDecimal256(137),
		VerifyingContract: "0x4bFb41d5B3570DeFd03C39a9A4D8fE6bD8ED413b",
	}

	OrderTypes = apitypes.Types{
		"EIP712Domain": []apitypes.Type{
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"Order": []apitypes.Type{
			{Name: "salt", Type: "uint256"},
			{Name: "maker", Type: "address"},
			{Name: "signer", Type: "address"},
			{Name: "taker", Type: "address"},
			{Name: "tokenId", Type: "uint256"},
			{Name: "makerAmount", Type: "uint256"},
			{Name: "takerAmount", Type: "uint256"},
			{Name: "expiration", Type: "uint256"},
			{Name: "nonce", Type: "uint256"},
			{Name: "feeRateBps", Type: "uint256"},
			{Name: "side", Type: "uint8"},
			{Name: "signatureType", Type: "uint8"},
		},
	}
)

type Order struct {
	Salt          *big.Int       `json:"salt"`
	Maker         common.Address `json:"maker"`
	Signer        common.Address `json:"signer"`
	Taker         common.Address `json:"taker"`
	TokenId       *big.Int       `json:"tokenId"`
	MakerAmount   *big.Int       `json:"makerAmount"`
	TakerAmount   *big.Int       `json:"takerAmount"`
	Expiration    *big.Int       `json:"expiration"`
	Nonce         *big.Int       `json:"nonce"`
	FeeRateBps    *big.Int       `json:"feeRateBps"`
	Side          uint8          `json:"side"`
	SignatureType uint8          `json:"signatureType"`
}

func formatBigInt(b *big.Int) string {
	if b == nil {
		return "0"
	}
	return b.String()
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		Salt        string `json:"salt"`
		TokenId     string `json:"tokenId"`
		MakerAmount string `json:"makerAmount"`
		TakerAmount string `json:"takerAmount"`
		Expiration  string `json:"expiration"`
		Nonce       string `json:"nonce"`
		FeeRateBps  string `json:"feeRateBps"`
		*Alias
	}{
		Salt:        formatBigInt(o.Salt),
		TokenId:     formatBigInt(o.TokenId),
		MakerAmount: formatBigInt(o.MakerAmount),
		TakerAmount: formatBigInt(o.TakerAmount),
		Expiration:  formatBigInt(o.Expiration),
		Nonce:       formatBigInt(o.Nonce),
		FeeRateBps:  formatBigInt(o.FeeRateBps),
		Alias:       (*Alias)(&o),
	})
}

// SignOrder signs a Polymarket Order using EIP-712 with the provided private key
func SignOrder(order Order, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types:       OrderTypes,
		PrimaryType: "Order",
		Domain:      PolymarketDomain,
		Message: map[string]interface{}{
			"salt":          (*math.HexOrDecimal256)(order.Salt),
			"maker":         order.Maker.Hex(),
			"signer":        order.Signer.Hex(),
			"taker":         order.Taker.Hex(),
			"tokenId":       (*math.HexOrDecimal256)(order.TokenId),
			"makerAmount":   (*math.HexOrDecimal256)(order.MakerAmount),
			"takerAmount":   (*math.HexOrDecimal256)(order.TakerAmount),
			"expiration":    (*math.HexOrDecimal256)(order.Expiration),
			"nonce":         (*math.HexOrDecimal256)(order.Nonce),
			"feeRateBps":    (*math.HexOrDecimal256)(order.FeeRateBps),
			"side":          math.NewHexOrDecimal256(int64(order.Side)),
			"signatureType": math.NewHexOrDecimal256(int64(order.SignatureType)),
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	return crypto.Sign(hash, privateKey)
}

func HashOrder(order Order) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types:       OrderTypes,
		PrimaryType: "Order",
		Domain:      PolymarketDomain,
		Message: map[string]interface{}{
			"salt":          (*math.HexOrDecimal256)(order.Salt),
			"maker":         order.Maker.Hex(),
			"signer":        order.Signer.Hex(),
			"taker":         order.Taker.Hex(),
			"tokenId":       (*math.HexOrDecimal256)(order.TokenId),
			"makerAmount":   (*math.HexOrDecimal256)(order.MakerAmount),
			"takerAmount":   (*math.HexOrDecimal256)(order.TakerAmount),
			"expiration":    (*math.HexOrDecimal256)(order.Expiration),
			"nonce":         (*math.HexOrDecimal256)(order.Nonce),
			"feeRateBps":    (*math.HexOrDecimal256)(order.FeeRateBps),
			"side":          math.NewHexOrDecimal256(int64(order.Side)),
			"signatureType": math.NewHexOrDecimal256(int64(order.SignatureType)),
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(typedData)
	return hash, err
}
