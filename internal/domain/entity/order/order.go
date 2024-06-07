package order

import (
	"math/big"
	"time"

	"github.com/r1der/epos/internal/domain/entity/pool"
	"github.com/r1der/epos/internal/domain/entity/project"
	"github.com/r1der/epos/internal/domain/values"
)

type Direction string

const (
	Buy  Direction = "buy"
	Sell Direction = "sell"
)

type Order struct {
	project        *project.Project
	pool           *pool.Pool
	address        string
	direction      Direction
	amountIn       values.Amount
	amountOut      values.Amount
	price          *big.Float
	transactionFee values.Amount
	createdAt      time.Time
}

func (ord *Order) Project() *project.Project { return ord.project }
func (ord *Order) Pool() *pool.Pool          { return ord.pool }
func (ord *Order) Address() string           { return ord.address }
func (ord *Order) Direction() Direction      { return ord.direction }
func (ord *Order) IsBuy() bool               { return ord.direction == Buy }
func (ord *Order) IsSell() bool              { return ord.direction == Sell }
func (ord *Order) AmountIn() values.Amount   { return ord.amountIn }
func (ord *Order) AmountOut() values.Amount  { return ord.amountOut }
func (ord *Order) FilledPrice() *big.Float   { return ord.price }
func (ord *Order) Fee() values.Amount        { return ord.transactionFee }
func (ord *Order) CreatedAt() time.Time      { return ord.createdAt }
