package erc20

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Manager interface {
	GetBalance(ctx context.Context, address common.Address) (*big.Int, error)
	GetTokenBalance(ctx context.Context, address, tokenAddress common.Address) (*big.Int, error)
}

type manager struct {
	cli ethclient.Client
}
