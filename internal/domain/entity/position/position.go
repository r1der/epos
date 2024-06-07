package position

import (
	"math/big"
	"time"

	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/project"
	"github.com/r1der/epos/internal/domain/values"
)

type Status string

const (
	Open   Status = "open"
	Closed Status = "closed"
)

type Position struct {
	project        *project.Project
	pool           *pool.Pool
	address        string
	lowerPrice     *big.Float
	upperPrice     *big.Float
	initialPrice   *big.Float
	liquidity      *big.Int
	inBaseAmount   values.Amount
	inQuoteAmount  values.Amount
	outBaseAmount  values.Amount
	outQuoteAmount values.Amount
	status         Status
	transactionFee values.Amount
	createdAt      time.Time
	closedAt       *time.Time

	// updatable values
	currentPrice            *big.Float
	currentBaseAmount       values.Amount
	currentQuoteAmount      values.Amount
	currentBaseAccruedFees  values.Amount
	currentQuoteAccruedFees values.Amount
}

func (p *Position) Project() *project.Project        { return p.project }
func (p *Position) Pool() *pool.Pool                 { return p.pool }
func (p *Position) Address() string                  { return p.address }
func (p *Position) LowerPrice() *big.Float           { return p.lowerPrice }
func (p *Position) UpperPrice() *big.Float           { return p.upperPrice }
func (p *Position) InitialPrice() *big.Float         { return p.initialPrice }
func (p *Position) Liquidity() *big.Int              { return p.liquidity }
func (p *Position) InputBaseAmount() values.Amount   { return p.inBaseAmount }
func (p *Position) InputQuoteAmount() values.Amount  { return p.inQuoteAmount }
func (p *Position) OutputBaseAmount() values.Amount  { return p.outBaseAmount }
func (p *Position) OutputQuoteAmount() values.Amount { return p.outQuoteAmount }
func (p *Position) Status() Status                   { return p.status }
func (p *Position) IsOpen() bool                     { return p.status == Open }
func (p *Position) IsClosed() bool                   { return p.status == Closed }
func (p *Position) TransactionFee() values.Amount    { return p.transactionFee }
func (p *Position) CreatedAt() time.Time             { return p.createdAt }
func (p *Position) ClosedAt() *time.Time             { return p.closedAt }

func (p *Position) CurrentPrice() *big.Float               { return p.currentPrice }
func (p *Position) CurrentBaseAmount() values.Amount       { return p.currentBaseAmount }
func (p *Position) CurrentQuoteAmount() values.Amount      { return p.currentQuoteAmount }
func (p *Position) CurrentBaseAccruedFees() values.Amount  { return p.currentBaseAccruedFees }
func (p *Position) CurrentQuoteAccruedFees() values.Amount { return p.currentQuoteAccruedFees }

func (p *Position) IsInRange() bool {
	if p.currentPrice.Cmp(p.lowerPrice) >= 0 && p.currentPrice.Cmp(p.upperPrice) <= 0 {
		return true
	}
	return false
}
