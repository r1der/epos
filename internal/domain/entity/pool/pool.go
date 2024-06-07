package pool

import (
	"fmt"
	"math/big"

	"github.com/r1der/epos/internal/domain/entity/token"
	"github.com/r1der/epos/internal/domain/values"
)

const (
	LowestFee values.Percent = 0.0001
	LowFee    values.Percent = 0.0005
	MediumFee values.Percent = 0.003
	HighFee   values.Percent = 0.01
)

type Pool struct {
	network   string
	protocol  string
	address   string
	fee       values.Percent
	pair      *token.Pair
	lastPrice *big.Float
}

func (p *Pool) Name() string {
	return fmt.Sprintf("[%s] %s: %s/%s [%f]",
		p.network, p.protocol, p.pair.BaseToken(), p.pair.QuoteToken(), p.fee)
}

func (p *Pool) Network() string       { return p.network }
func (p *Pool) Protocol() string      { return p.protocol }
func (p *Pool) Address() string       { return p.address }
func (p *Pool) Fee() values.Percent   { return p.fee }
func (p *Pool) Pair() *token.Pair     { return p.pair }
func (p *Pool) LastPrice() *big.Float { return p.lastPrice }

func (p *Pool) updatePrice(price *big.Float) {
	p.lastPrice = price
}
