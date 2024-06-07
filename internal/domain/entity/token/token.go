package token

import (
	"math"
	"math/big"
)

type Ticker string

func (t Ticker) String() string { return string(t) }

type Token struct {
	network  string
	address  string
	ticker   Ticker
	decimals int
}

func (t *Token) Network() string { return t.network }
func (t *Token) Address() string { return t.address }
func (t *Token) Ticker() Ticker  { return t.ticker }
func (t *Token) Decimals() int   { return t.decimals }
func (t *Token) String() string  { return t.ticker.String() }

func (t *Token) ToBaseValue(v *big.Float) *big.Int {
	baseValue, _ := new(big.Float).
		Mul(v, new(big.Float).SetFloat64(math.Pow10(t.decimals))).
		Int(nil)
	return baseValue
}

func (t *Token) ToHumanValue(v *big.Int) *big.Float {
	return new(big.Float).
		Quo(new(big.Float).SetInt(v), new(big.Float).SetFloat64(math.Pow10(t.decimals)))
}

func (t *Token) Eq(t2 *Token) bool {
	if t.network == t2.network && t.address == t2.address {
		return true
	}
	return false
}

func New(network, address string, ticker Ticker, decimals int) *Token {
	return &Token{
		network:  network,
		address:  address,
		ticker:   ticker,
		decimals: decimals,
	}
}
