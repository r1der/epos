package ports

import (
	"context"
	"math/big"

	"github.com/r1der/epos/internal/domain/entity/token"
	"github.com/r1der/epos/internal/domain/values"
)

type LiquidityManager interface {
	IncreaseLiquidity(ctx context.Context, in *IncreaseLiquidityInput) (*IncreaseLiquidityOutput, error)
	DecreaseLiquidity(ctx context.Context, in *DecreaseLiquidityInput) (*DecreaseLiquidityOutput, error)
	GetPosition(ctx context.Context, in *GetPositionInput) (*GetPositionOutput, error)
}

type IncreaseLiquidityInput struct {
	Network     string
	Protocol    string
	PoolAddress string
	Fee         values.Percent
	Pair        *token.Pair
	LowerPrice  *big.Float
	UpperPrice  *big.Float
	BaseAmount  values.Amount
	QuoteAmount values.Amount
}

type IncreaseLiquidityOutput struct {
	Address        string
	Liquidity      *big.Int
	BaseAmount     values.Amount
	QuoteAmount    values.Amount
	TransactionFee *big.Int
}

type DecreaseLiquidityInput struct {
	Network         string
	Protocol        string
	PoolAddress     string
	Fee             values.Percent
	Pair            *token.Pair
	PositionAddress string
	Liquidity       *big.Int
	BaseMaxAmount   values.Amount
	QuoteMaxAmount  values.Amount
}

type DecreaseLiquidityOutput struct {
	Address        string
	Liquidity      *big.Int
	BaseAmount     values.Amount
	QuoteAmount    values.Amount
	TransactionFee *big.Int
}

type GetPositionInput struct {
	Network         string
	Protocol        string
	PoolAddress     string
	Fee             values.Percent
	Pair            *token.Pair
	PositionAddress string
}

type GetPositionOutput struct {
	CurrentPrice     *big.Float
	Liquidity        *big.Int
	BaseAmount       values.Amount
	QuoteAmount      values.Amount
	BaseAccruedFees  values.Amount
	QuoteAccruedFees values.Amount
}
