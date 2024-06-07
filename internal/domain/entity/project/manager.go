package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/wallet"
	"github.com/r1der/epos/internal/domain/ports"
	"github.com/r1der/epos/internal/domain/values"
)

var (
	ErrInvestmentsNotEnough = errors.New("investments not enough")
)

type Manager interface {
	New(ctx context.Context, in *NewProjectInput) (*Project, error)
	Deactivate(ctx context.Context, proj *Project, reason InactiveReason) error
	UpdateWorth(ctx context.Context, proj *Project, worth values.Amount) error
}

type manager struct {
	repo    Repository
	balance ports.Balance
}

func NewManager(repo Repository, balance ports.Balance) Manager {
	return &manager{
		repo:    repo,
		balance: balance,
	}
}

type NewProjectInput struct {
	Wallet          *wallet.Wallet
	Pool            *pool.Pool
	Name            string
	Investments     values.Amount
	TakeProfit      values.Percent
	StopLoss        values.Percent
	RangeVolatility values.Percent
	Slippage        values.Percent
	ActivePositions int
}

// New creates a new smart-pool strategy
func (svc *manager) New(ctx context.Context, in *NewProjectInput) (*Project, error) {
	// we check if there are funds for investment in the strategy
	bal, err := svc.balance.Get(ctx, in.Wallet, in.Investments.Token())
	if err != nil {
		return nil, fmt.Errorf("check funds for investment: %w", err)
	}

	if bal.Value().Cmp(in.Investments.Value()) < 0 {
		return nil, ErrInvestmentsNotEnough
	}

	proj := &Project{
		id:              uuid.New(),
		wallet:          in.Wallet,
		pool:            in.Pool,
		name:            in.Name,
		investments:     in.Investments,
		takeProfit:      in.TakeProfit,
		stopLoss:        in.StopLoss,
		rangeVolatility: in.RangeVolatility,
		slippage:        in.Slippage,
		status:          Active,
		inactiveReason:  EmptyReason,
		currentValue:    in.Investments,
		createdAt:       time.Now(),
	}

	if err = svc.repo.Save(ctx, proj); err != nil {
		return nil, fmt.Errorf("save project after create: %w", err)
	}
	return proj, nil
}

// Deactivate makes the project as inactive
func (svc *manager) Deactivate(ctx context.Context, proj *Project, reason InactiveReason) error {
	proj.status = Inactive
	proj.inactiveReason = reason
	if err := svc.repo.Save(ctx, proj); err != nil {
		return fmt.Errorf("save project after deactivate: %w", err)
	}
	return nil
}

// UpdateWorth updates a current project worth
func (svc *manager) UpdateWorth(ctx context.Context, proj *Project, worth values.Amount) error {
	proj.currentValue = worth
	if err := svc.repo.Save(ctx, proj); err != nil {
		return fmt.Errorf("save project after update worth: %w", err)
	}
	return nil
}
