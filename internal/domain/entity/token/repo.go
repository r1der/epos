package token

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("token not found")
)

type Repository interface {
	FindOne(context.Context, Filter) (*Token, error)
	Find(context.Context, Filter) ([]*Token, error)
	Save(context.Context, *Token) error
}

type Filter struct {
	Networks  []string
	Addresses []string
	Tickers   []Ticker
}

type OrderBy string
