package values

import (
	"math/big"
)

type Percent float64

func (p Percent) Value() float64    { return float64(p) }
func (p Percent) Float() *big.Float { return big.NewFloat(float64(p)) }

func NewPercent(percent float64) Percent {
	return Percent(percent)
}
