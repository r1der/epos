package wallet

import (
	"time"

	"github.com/r1der/epos/internal/domain/entity/token"
)

type Wallet struct {
	name        string
	network     string
	address     string
	privateKey  string
	nativeToken *token.Token
	createdAt   time.Time
}

func (w *Wallet) Name() string              { return w.name }
func (w *Wallet) NetworkId() string         { return w.network }
func (w *Wallet) Address() string           { return w.address }
func (w *Wallet) PrivateKey() string        { return w.privateKey }
func (w *Wallet) NativeToken() *token.Token { return w.nativeToken }
func (w *Wallet) CreatedAt() time.Time      { return w.createdAt }
