package ports

import (
	"context"
	"math/big"

	"github.com/r1der/epos/internal/domain/values"
)

type Router interface {
	Swap(ctx context.Context, in *SwapInput) (*SwapOutput, error)
}

type SwapInput struct {
	Network     string
	Protocol    string
	PoolAddress string
	Fee         values.Percent
	AmountIn    values.Amount
	AmountOut   values.Amount
}

type SwapOutput struct {
	Address        string
	AmountIn       values.Amount
	AmountOut      values.Amount
	FilledPrice    *big.Float
	TransactionFee *big.Int
}
