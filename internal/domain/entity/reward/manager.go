package reward

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/r1der/epos/internal/domain/entity/position"
	"github.com/r1der/epos/internal/domain/values"
)

type Manager interface {
	Add(context.Context, *position.Position, ...values.Amount) ([]*Reward, error)
	GetPositionRewards(ctx context.Context, pos *position.Position) ([]*Reward, error)
}

type manager struct {
	repo Repository
}

func NewManager(repo Repository) Manager {
	return &manager{repo: repo}
}

// Add creates new rewards for the position
func (svc *manager) Add(ctx context.Context, pos *position.Position, aa ...values.Amount) ([]*Reward, error) {
	if len(aa) == 0 {
		return nil, nil
	}

	rewards := make([]*Reward, 0)
	for _, amount := range aa {
		rewards = append(rewards, &Reward{
			id:        uuid.New(),
			pos:       pos,
			amount:    amount,
			createdAt: time.Now(),
		})
	}

	if err := svc.repo.Save(ctx, rewards...); err != nil {
		return nil, fmt.Errorf("save rewards in repo: %w", err)
	}

	return rewards, nil
}

// GetPositionRewards gets all rewards for the position
func (svc *manager) GetPositionRewards(ctx context.Context, pos *position.Position) ([]*Reward, error) {
	return svc.repo.Find(ctx, Filter{Positions: []*position.Position{pos}})
}
