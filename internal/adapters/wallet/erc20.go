package wallet

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const erc20ABI = `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

type ERC20Token struct {
	client   *ethclient.Client
	address  common.Address
	parsedABI abi.ABI
}

func NewERC20Token(client *ethclient.Client, address common.Address) (*ERC20Token, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, err
	}
	return &ERC20Token{
		client:    client,
		address:   address,
		parsedABI: parsed,
	}, nil
}

func (t *ERC20Token) BalanceOf(ctx context.Context, owner common.Address) (*big.Int, error) {
	data, err := t.parsedABI.Pack("balanceOf", owner)
	if err != nil {
		return nil, err
	}
	
	msg := ethereum.CallMsg{
		To:   &t.address,
		Data: data,
	}
	
	res, err := t.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}
	
	out, err := t.parsedABI.Unpack("balanceOf", res)
	if err != nil || len(out) == 0 {
		return nil, err
	}
	
	return out[0].(*big.Int), nil
}

func (t *ERC20Token) Allowance(ctx context.Context, owner, spender common.Address) (*big.Int, error) {
	data, err := t.parsedABI.Pack("allowance", owner, spender)
	if err != nil {
		return nil, err
	}
	
	msg := ethereum.CallMsg{
		To:   &t.address,
		Data: data,
	}
	
	res, err := t.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}
	
	out, err := t.parsedABI.Unpack("allowance", res)
	if err != nil || len(out) == 0 {
		return nil, err
	}
	
	return out[0].(*big.Int), nil
}

func (t *ERC20Token) BuildApproveTx(ctx context.Context, spender common.Address, amount *big.Int, nonce uint64, gasPrice *big.Int) (*types.Transaction, error) {
	data, err := t.parsedABI.Pack("approve", spender, amount)
	if err != nil {
		return nil, err
	}
	
	gasLimit, err := t.client.EstimateGas(ctx, ethereum.CallMsg{
		To:   &t.address,
		Data: data,
	})
	if err != nil {
		gasLimit = 100000 // safe fallback
	}

	tx := types.NewTransaction(nonce, t.address, big.NewInt(0), gasLimit, gasPrice, data)
	return tx, nil
}
