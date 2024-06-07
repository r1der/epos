package project

import (
	"time"

	"github.com/google/uuid"

	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/wallet"
	"github.com/r1der/epos/internal/domain/values"
)

type (
	Status         string
	InactiveReason string
)

const (
	Active   Status = "active"
	Inactive Status = "inactive"

	StopLoss       InactiveReason = "stop-loss"
	TakeProfit     InactiveReason = "take-profit"
	NotEnoughFunds InactiveReason = "not-enough-funds"
	NotEnoughGas   InactiveReason = "not-enough-gas"
	EmptyReason    InactiveReason = ""
)

type Project struct {
	id              uuid.UUID
	wallet          *wallet.Wallet
	pool            *pool.Pool
	name            string
	investments     values.Amount
	takeProfit      values.Percent
	stopLoss        values.Percent
	rangeVolatility values.Percent
	slippage        values.Percent
	activePositions int
	currentValue    values.Amount
	status          Status
	inactiveReason  InactiveReason
	createdAt       time.Time
}

func (p *Project) ID() uuid.UUID                   { return p.id }
func (p *Project) Wallet() *wallet.Wallet          { return p.wallet }
func (p *Project) Pool() *pool.Pool                { return p.pool }
func (p *Project) Name() string                    { return p.name }
func (p *Project) Investments() values.Amount      { return p.investments }
func (p *Project) TakeProfit() values.Percent      { return p.takeProfit }
func (p *Project) StopLoss() values.Percent        { return p.stopLoss }
func (p *Project) RangeVolatility() values.Percent { return p.rangeVolatility }
func (p *Project) Slippage() values.Percent        { return p.slippage }
func (p *Project) ActivePositions() int            { return p.activePositions }
func (p *Project) Status() Status                  { return p.status }
func (p *Project) IsActive() bool                  { return p.status == Active }
func (p *Project) IsInactive() bool                { return p.status == Inactive }
func (p *Project) InactiveReason() InactiveReason  { return p.inactiveReason }
func (p *Project) CurrentValue() values.Amount     { return p.currentValue }
func (p *Project) CreatedAt() time.Time            { return p.createdAt }
