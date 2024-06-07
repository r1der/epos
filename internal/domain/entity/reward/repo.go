package reward

import (
	"context"
	"errors"

	"github.com/r1der/epos/internal/domain/entity/position"
	"github.com/r1der/epos/internal/domain/entity/token"
)

var (
	ErrNotFound = errors.New("reward not found")
)

type Repository interface {
	FindOne(context.Context, Filter) (*Reward, error)
	Find(context.Context, Filter) ([]*Reward, error)
	Save(context.Context, ...*Reward) error
}

type Filter struct {
	Positions []*position.Position
	Tokens    []*token.Token
}

type OrderBy string
