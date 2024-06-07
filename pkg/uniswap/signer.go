package uniswap

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

type Signer interface {
	PrivateKey() *ecdsa.PrivateKey
	PublicKey() *ecdsa.PublicKey
	Address() common.Address
}
