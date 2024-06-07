package project

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/wallet"
)

var (
	ErrNotFound = errors.New("project not found")
)

type Repository interface {
	FindOne(context.Context, Filter) (*Project, error)
	Find(context.Context, Filter) ([]*Project, error)
	Save(context.Context, *Project) error
}

type Filter struct {
	Ids      []uuid.UUID
	Wallets  []*wallet.Wallet
	Pools    []*pool.Pool
	Statuses []Status
	Reasons  []InactiveReason
}

type OrderBy string
