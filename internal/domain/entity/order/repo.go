package order

import (
	"context"
	"errors"

	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/project"
	"github.com/r1der/epos/internal/domain/entity/token"
)

var (
	ErrNotFound = errors.New("order not found")
)

type Repository interface {
	FindOne(context.Context, Filter) (*Order, error)
	Find(context.Context, Filter) ([]*Order, error)
	Save(context.Context, *Order) error
}

type Filter struct {
	Projects   []*project.Project
	Pools      []*pool.Pool
	Addresses  []string
	Directions []Direction
	TokensIn   []*token.Token
	TokensOut  []*token.Token
}

type SortBy string
