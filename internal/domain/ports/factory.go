package ports

import (
	"context"
	"math/big"

	"github.com/r1der/epos/internal/domain/entity/token"
	"github.com/r1der/epos/internal/domain/values"
)

type Pool struct {
	Address   string
	LastPrice *big.Float
	Liquidity *big.Int
}

type Factory interface {
	FindPool(ctx context.Context, network, protocol string, pair *token.Pair, fee values.Percent) (*Pool, error)
	GetPool(ctx context.Context, network, protocol, address string) (*Pool, error)
	CalculateRange(ctx context.Context, in *CalculateRangeInput) (*CalculateRangeOutput, error)
	CalculateAmounts(ctx context.Context, in *CalculateAmountsInput) (*CalculateAmountsOutput, error)
}

type CalculateRangeInput struct {
	Network         string
	Protocol        string
	PoolAddress     string
	BaseVolatility  values.Percent
	QuoteVolatility values.Percent
}

type CalculateRangeOutput struct {
	Network     string
	Protocol    string
	PoolAddress string
	LastPrice   *big.Float
	LowerPrice  *big.Float
	UpperPrice  *big.Float
}

type CalculateAmountsInput struct {
	InitialPrice *big.Float
	LowerPrice   *big.Float
	UpperPrice   *big.Float
	BaseAmount   values.Amount
	QuoteAmount  values.Amount
}

type CalculateAmountsOutput struct {
	Liquidity   *big.Int
	BaseAmount  values.Amount
	QuoteAmount values.Amount
}
