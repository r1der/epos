package position

import (
	"context"
	"errors"

	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/project"
)

var (
	ErrNotFound = errors.New("position not found")
)

type Repository interface {
	FindOne(context.Context, Filter) (*Position, error)
	Find(context.Context, Filter) ([]*Position, error)
	Save(context.Context, *Position) error
}

type Filter struct {
	Projects  []*project.Project
	Pools     []*pool.Pool
	Addresses []string
	Statuses  []Status
}

type OrderBy string
