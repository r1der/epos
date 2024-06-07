package wallet

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("wallet not found")
)

type Repository interface {
	FindOne(context.Context, Filter) (*Wallet, error)
	FindO(context.Context, Filter) ([]*Wallet, error)
	Save(context.Context, *Wallet) error
}

type Filter struct {
	Networks  []string
	Addresses []string
}

type OrderBy string
