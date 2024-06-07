package pool

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/r1der/epos/internal/domain/entity/token"
	"github.com/r1der/epos/internal/domain/ports"
	"github.com/r1der/epos/internal/domain/values"
)

type Manager interface {
	Get(ctx context.Context, network, protocol string, pair *token.Pair, fee values.Percent) (*Pool, error)
	CalculatePositionRange(ctx context.Context, p *Pool, volatility values.Percent) (*Range, error)
	CalculatePositionAmounts(ctx context.Context, pricesRange *Range, baseAmount, quoteAmount values.Amount) (*Amounts, error)
}

type manager struct {
	repo    Repository
	factory ports.Factory
}

func NewManager(repo Repository, factory ports.Factory) Manager {
	return &manager{
		repo:    repo,
		factory: factory,
	}
}

// Get gets a pool
func (svc *manager) Get(ctx context.Context, network, protocol string, pair *token.Pair, fee values.Percent) (*Pool, error) {
	p, err := svc.repo.FindOne(ctx, Filter{
		Networks:    []string{network},
		Protocols:   []string{protocol},
		BaseTokens:  []*token.Token{pair.BaseToken()},
		QuoteTokens: []*token.Token{pair.QuoteToken()},
		Fees:        []values.Percent{fee},
	})
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("find a pool in repo: %w", err)
	}

	var data *ports.Pool

	if p != nil {
		data, err = svc.factory.GetPool(ctx, p.network, p.protocol, p.address)
		if err != nil {
			return nil, fmt.Errorf("factory: get pool: %w", err)
		}

		p.updatePrice(data.LastPrice)
		if err = svc.repo.Save(ctx, p); err != nil {
			return nil, fmt.Errorf("save pool after update price: %w", err)
		}
		return p, nil
	}

	data, err = svc.factory.FindPool(ctx, network, protocol, pair, fee)
	if err != nil {
		return nil, fmt.Errorf("factory: find pool: %w", err)
	}

	p = &Pool{
		network:   network,
		protocol:  protocol,
		address:   data.Address,
		fee:       fee,
		pair:      pair,
		lastPrice: data.LastPrice,
	}
	if err = svc.repo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("save pool after create: %w", err)
	}

	return p, nil
}

type Range struct {
	InitialPrice *big.Float
	LowerPrice   *big.Float
	UpperPrice   *big.Float
}

// CalculatePositionRange calculates the price range for a position based on volatility
func (svc *manager) CalculatePositionRange(ctx context.Context, p *Pool, volatility values.Percent) (*Range, error) {
	data, err := svc.factory.CalculateRange(ctx, &ports.CalculateRangeInput{
		Network:         p.network,
		Protocol:        p.protocol,
		PoolAddress:     p.address,
		BaseVolatility:  volatility,
		QuoteVolatility: volatility,
	})
	if err != nil {
		return nil, fmt.Errorf("factory: calculate range: %w", err)
	}
	p.lastPrice = data.LastPrice
	return &Range{
		InitialPrice: data.LastPrice,
		LowerPrice:   data.LowerPrice,
		UpperPrice:   data.UpperPrice,
	}, nil
}

type Amounts struct {
	Liquidity   *big.Int
	BaseAmount  values.Amount
	QuoteAmount values.Amount
}

// CalculatePositionAmounts calculates the asset amounts for a position based on prices range
func (svc *manager) CalculatePositionAmounts(ctx context.Context, pricesRange *Range, baseAmount, quoteAmount values.Amount) (*Amounts, error) {
	data, err := svc.factory.CalculateAmounts(ctx, &ports.CalculateAmountsInput{
		InitialPrice: pricesRange.InitialPrice,
		LowerPrice:   pricesRange.LowerPrice,
		UpperPrice:   pricesRange.UpperPrice,
		BaseAmount:   baseAmount,
		QuoteAmount:  quoteAmount,
	})
	if err != nil {
		return nil, fmt.Errorf("factory: calculate amounts: %w", err)
	}
	return &Amounts{
		Liquidity:   data.Liquidity,
		BaseAmount:  data.BaseAmount,
		QuoteAmount: data.QuoteAmount,
	}, nil
}
