package position

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/r1der/epos/internal/domain/entity/project"
	"github.com/r1der/epos/internal/domain/ports"
	"github.com/r1der/epos/internal/domain/values"
)

type Manager interface {
	Open(ctx context.Context, in *OpenPositionInput) (*Position, error)
	Close(ctx context.Context, pos *Position) error
	Actualize(ctx context.Context, pos *Position) error
	CollectRewards(ctx context.Context, pos *Position) ([]values.Amount, error)

	GetOpenPositions(ctx context.Context, proj *project.Project) ([]*Position, error)
}

type manager struct {
	repo             Repository
	liquidityManager ports.LiquidityManager
}

func NewManager(repo Repository, liquidityManager ports.LiquidityManager) Manager {
	return &manager{
		repo:             repo,
		liquidityManager: liquidityManager,
	}
}

// GetOpenPositions gets open positions for selected project
func (svc *manager) GetOpenPositions(ctx context.Context, proj *project.Project) ([]*Position, error) {
	pp, err := svc.repo.Find(ctx, Filter{Projects: []*project.Project{proj}, Statuses: []Status{Open}})
	if err != nil {
		return nil, fmt.Errorf("find open positions in repo: %w", err)
	}
	return pp, nil
}

type OpenPositionInput struct {
	Project     *project.Project
	InitPrice   *big.Float
	LowerPrice  *big.Float
	UpperPrice  *big.Float
	BaseAmount  values.Amount
	QuoteAmount values.Amount
}

// Open creates a new position
func (svc *manager) Open(ctx context.Context, in *OpenPositionInput) (*Position, error) {
	logrus.Debugf("start of opening a new position")

	p := in.Project.Pool()
	data, err := svc.liquidityManager.IncreaseLiquidity(ctx, &ports.IncreaseLiquidityInput{
		Network:     p.Network(),
		Protocol:    p.Protocol(),
		PoolAddress: p.Address(),
		Pair:        p.Pair(),
		Fee:         p.Fee(),
		LowerPrice:  in.LowerPrice,
		UpperPrice:  in.UpperPrice,
		BaseAmount:  in.BaseAmount,
		QuoteAmount: in.QuoteAmount,
	})
	if err != nil {
		return nil, fmt.Errorf("liquidity manager: increase liquidity: %w", err)
	}

	pos := &Position{
		project:                 in.Project,
		pool:                    p,
		address:                 data.Address,
		lowerPrice:              in.LowerPrice,
		upperPrice:              in.UpperPrice,
		initialPrice:            in.InitPrice,
		liquidity:               data.Liquidity,
		inBaseAmount:            data.BaseAmount,
		inQuoteAmount:           data.QuoteAmount,
		status:                  Open,
		transactionFee:          values.NewAmount(in.Project.Wallet().NativeToken(), data.TransactionFee),
		createdAt:               time.Now(),
		currentPrice:            in.InitPrice,
		currentBaseAmount:       data.BaseAmount,
		currentQuoteAmount:      data.QuoteAmount,
		outBaseAmount:           values.NewAmount(data.BaseAmount.Token(), 0),
		outQuoteAmount:          values.NewAmount(data.QuoteAmount.Token(), 0),
		currentBaseAccruedFees:  values.NewAmount(data.BaseAmount.Token(), 0),
		currentQuoteAccruedFees: values.NewAmount(data.QuoteAmount.Token(), 0),
	}

	if err = svc.repo.Save(ctx, pos); err != nil {
		return nil, fmt.Errorf("save position in repo after create: %w", err)
	}

	return pos, nil
}

// Actualize updates information about position
func (svc *manager) Actualize(ctx context.Context, pos *Position) error {
	data, err := svc.liquidityManager.GetPosition(ctx, &ports.GetPositionInput{})
	if err != nil {
		return fmt.Errorf("liquidity manager: get position: %w", err)
	}

	pos.currentPrice = data.CurrentPrice
	pos.currentBaseAmount = data.BaseAmount
	pos.currentQuoteAmount = data.QuoteAmount
	pos.currentBaseAccruedFees = data.BaseAccruedFees
	pos.currentQuoteAccruedFees = data.QuoteAccruedFees

	if err = svc.repo.Save(ctx, pos); err != nil {
		return fmt.Errorf("save position in repo after actualize: %w", err)
	}

	return nil
}

// Close closes position (remove liquidity)
func (svc *manager) Close(ctx context.Context, pos *Position) error {
	// нужно учесть slippage для получения активов
	baseAmountSlippage := pos.currentBaseAmount.Mul(pos.Project().Slippage())
	baseAmountWithSlippage := pos.currentBaseAmount.Sub(baseAmountSlippage)
	quoteAmountSlippage := pos.currentQuoteAmount.Mul(pos.Project().Slippage())
	quoteAmountWithSlippage := pos.currentQuoteAmount.Sub(quoteAmountSlippage)

	data, err := svc.liquidityManager.DecreaseLiquidity(ctx, &ports.DecreaseLiquidityInput{
		Network:         pos.Project().Pool().Network(),
		Protocol:        pos.Project().Pool().Protocol(),
		PoolAddress:     pos.Project().Pool().Address(),
		Fee:             pos.Project().Pool().Fee(),
		Pair:            pos.Project().Pool().Pair(),
		PositionAddress: pos.Address(),
		Liquidity:       pos.liquidity,
		BaseMaxAmount:   baseAmountWithSlippage,
		QuoteMaxAmount:  quoteAmountWithSlippage,
	})
	if err != nil {
		return fmt.Errorf("liquidity manager: decrease liquidity: %w", err)
	}

	pos.outBaseAmount = data.BaseAmount
	pos.outQuoteAmount = data.QuoteAmount

	if err = svc.repo.Save(ctx, pos); err != nil {
		return fmt.Errorf("save position in repo after close: %w", err)
	}

	return nil
}

// CollectFees collects position fees
func (svc *manager) CollectRewards(ctx context.Context, pos *Position) ([]values.Amount, error) {
	return nil, nil
}
