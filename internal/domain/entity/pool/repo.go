package pool

import (
	"context"
	"errors"

	"github.com/r1der/epos/internal/domain/entity/token"
	"github.com/r1der/epos/internal/domain/values"
)

var (
	ErrNotFound = errors.New("pool not found")
)

type Repository interface {
	FindOne(context.Context, Filter) (*Pool, error)
	Find(context.Context, Filter) ([]*Pool, error)
	Save(context.Context, *Pool) error
}

type Filter struct {
	Networks    []string
	Protocols   []string
	Addresses   []string
	BaseTokens  []*token.Token
	QuoteTokens []*token.Token
	Fees        []values.Percent
}

type OrderBy string
