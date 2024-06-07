package order

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/r1der/epos/internal/domain/entity/project"
	"github.com/r1der/epos/internal/domain/ports"
	"github.com/r1der/epos/internal/domain/values"
)

type Manager interface {
	New(ctx context.Context, in *NewOrderInput) (*Order, error)
}

type manager struct {
	repo   Repository
	router ports.Router
}

func NewManager(repo Repository, router ports.Router) Manager {
	return &manager{
		repo:   repo,
		router: router,
	}
}

type NewOrderInput struct {
	Project   *project.Project
	AmountIn  values.Amount
	AmountOut values.Amount
	Price     *big.Float
}

func (svc *manager) New(ctx context.Context, in *NewOrderInput) (*Order, error) {
	data, err := svc.router.Swap(ctx, &ports.SwapInput{
		Network:     in.Project.Pool().Network(),
		Protocol:    in.Project.Pool().Protocol(),
		PoolAddress: in.Project.Pool().Address(),
		Fee:         in.Project.Pool().Fee(),
		AmountIn:    in.AmountIn,
		AmountOut:   in.AmountOut,
	})
	if err != nil {
		return nil, fmt.Errorf("router: swap: %w", err)
	}

	var direction Direction
	if in.Project.Pool().Pair().BaseToken().Eq(in.AmountIn.Token()) {
		direction = Sell
	} else {
		direction = Buy
	}

	ord := &Order{
		project:        in.Project,
		pool:           in.Project.Pool(),
		address:        data.Address,
		direction:      direction,
		amountIn:       data.AmountIn,
		amountOut:      data.AmountOut,
		price:          data.FilledPrice,
		transactionFee: values.NewAmount(in.Project.Wallet().NativeToken(), data.TransactionFee),
		createdAt:      time.Now(),
	}

	if err = svc.repo.Save(ctx, ord); err != nil {
		return nil, fmt.Errorf("save order in repo after create: %w", err)
	}

	return ord, nil
}
