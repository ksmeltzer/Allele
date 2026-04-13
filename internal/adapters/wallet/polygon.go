package wallet

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type PolygonWallet struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewPolygonWallet(pk *ecdsa.PrivateKey) *PolygonWallet {
	address := crypto.PubkeyToAddress(pk.PublicKey)
	return &PolygonWallet{
		privateKey: pk,
		address:    address,
	}
}

func (w *PolygonWallet) Address() string {
	return w.address.Hex()
}

func (w *PolygonWallet) SignMessage(payload []byte) ([]byte, error) {
	sig, err := crypto.Sign(payload, w.privateKey)
	if err != nil {
		return nil, err
	}
	if len(sig) == 65 {
		sig[64] += 27
	}
	return sig, nil
}
