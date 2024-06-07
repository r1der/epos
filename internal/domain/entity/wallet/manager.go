package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/r1der/epos/internal/domain/entity/token"
	"github.com/r1der/epos/internal/domain/ports"
)

type Manager interface {
	New(ctx context.Context, in *NewWalletInput) (*Wallet, error)
	Get(ctx context.Context, network, address string) (*Wallet, error)
}

type manager struct {
	repo          Repository
	addrGenerator ports.WalletAddressGenerator
}

func NewManager(repo Repository, addrGenerator ports.WalletAddressGenerator) Manager {
	return &manager{
		repo:          repo,
		addrGenerator: addrGenerator,
	}
}

type NewWalletInput struct {
	Name        string
	Network     string
	PrivateKey  string
	NativeToken *token.Token
}

// New creates a new wallet in selected network
func (svc *manager) New(ctx context.Context, in *NewWalletInput) (*Wallet, error) {
	addr, err := svc.addrGenerator.Generate(ctx, in.Network, in.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("generate wallet address: %w", err)
	}
	return &Wallet{
		name:        in.Name,
		network:     in.Network,
		address:     addr,
		privateKey:  in.PrivateKey,
		nativeToken: in.NativeToken,
		createdAt:   time.Now(),
	}, nil
}

// Get finds a wallet
func (svc *manager) Get(ctx context.Context, network, address string) (*Wallet, error) {
	return svc.repo.FindOne(ctx, Filter{Networks: []string{network}, Addresses: []string{address}})
}
