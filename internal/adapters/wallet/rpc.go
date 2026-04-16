package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type RPCManager struct {
	client    *ethclient.Client
	wallet    *PolygonWallet
	network   string
	usdcAddr  common.Address
	ctfAddr   common.Address
	usdcToken *ERC20Token
}

func NewRPCManager(rpcURL, network string, pk *ecdsa.PrivateKey) (*RPCManager, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	wallet := NewPolygonWallet(pk)

	// Default to Polygon Mainnet
	usdcAddr := common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174") // USDC.e
	ctfAddr := common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8fE6bD8ED413b")  // CTF Exchange

	if network == "Polygon Amoy Testnet" {
		usdcAddr = common.HexToAddress("0x41E94eb019C0762f9Bfcf9Fb1e58725BfB0e7582") // Amoy USDC
		ctfAddr = common.HexToAddress("0x4bFb41d5B3570DeFd03C39a9A4D8fE6bD8ED413b")  // CTF Exchange Amoy
	}

	usdcToken, err := NewERC20Token(client, usdcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create USDC token wrapper: %w", err)
	}

	return &RPCManager{
		client:    client,
		wallet:    wallet,
		network:   network,
		usdcAddr:  usdcAddr,
		ctfAddr:   ctfAddr,
		usdcToken: usdcToken,
	}, nil
}

func (r *RPCManager) GetBalances(ctx context.Context) (*big.Int, *big.Int, error) {
	maticBalance, err := r.client.BalanceAt(ctx, r.wallet.address, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get MATIC balance: %w", err)
	}

	usdcBalance, err := r.usdcToken.BalanceOf(ctx, r.wallet.address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get USDC balance: %w", err)
	}

	return maticBalance, usdcBalance, nil
}

func (r *RPCManager) CheckAndApproveUSDC(ctx context.Context) error {
	allowance, err := r.usdcToken.Allowance(ctx, r.wallet.address, r.ctfAddr)
	if err != nil {
		return fmt.Errorf("failed to check USDC allowance: %w", err)
	}

	minAllowance := big.NewInt(10000000)

	if allowance.Cmp(minAllowance) < 0 {
		log.Printf("USDC allowance low (%v). Approving CTF Exchange...", allowance)

		maticBalance, err := r.client.BalanceAt(ctx, r.wallet.address, nil)
		if err != nil {
			return fmt.Errorf("failed to check MATIC balance before approval: %w", err)
		}
		if maticBalance.Cmp(big.NewInt(0)) == 0 {
			return fmt.Errorf("insufficient MATIC for gas. Cannot approve USDC until wallet is funded")
		}

		nonce, err := r.client.PendingNonceAt(ctx, r.wallet.address)
		if err != nil {
			return fmt.Errorf("failed to get nonce: %w", err)
		}

		gasPrice, err := r.client.SuggestGasPrice(ctx)
		if err != nil {
			return fmt.Errorf("failed to suggest gas price: %w", err)
		}

		maxAmount := new(big.Int)
		maxAmount.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)

		tx, err := r.usdcToken.BuildApproveTx(ctx, r.ctfAddr, maxAmount, nonce, gasPrice)
		if err != nil {
			return fmt.Errorf("failed to build approve tx: %w", err)
		}

		chainID, err := r.client.ChainID(ctx)
		if err != nil {
			return fmt.Errorf("failed to get chain ID: %w", err)
		}

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), r.wallet.privateKey)
		if err != nil {
			return fmt.Errorf("failed to sign approve tx: %w", err)
		}

		err = r.client.SendTransaction(ctx, signedTx)
		if err != nil {
			return fmt.Errorf("failed to broadcast approve tx: %w", err)
		}

		log.Printf("Sent USDC approve transaction: %s", signedTx.Hash().Hex())
	}

	return nil
}
