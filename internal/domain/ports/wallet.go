package ports

import (
	"context"

	"github.com/r1der/epos/internal/domain/entity/token"
	"github.com/r1der/epos/internal/domain/entity/wallet"
	"github.com/r1der/epos/internal/domain/values"
)

type Balance interface {
	Get(ctx context.Context, wa *wallet.Wallet, token *token.Token) (values.Amount, error)
}

type WalletAddressGenerator interface {
	Generate(ctx context.Context, network, privateKey string) (string, error)
}
