package values

import (
	"fmt"
	"math/big"

	"github.com/r1der/epos/internal/domain/entity/token"
)

type Amount struct {
	token *token.Token
	value *big.Int
}

func (a Amount) Value() *big.Int     { return a.value }
func (a Amount) Token() *token.Token { return a.token }

func (a Amount) IsZero() bool {
	if a.value.Cmp(big.NewInt(0)) == 0 {
		return true
	}
	return false
}

func (a Amount) Add(a2 Amount) Amount {
	if !a.token.Eq(a2.token) {
		panic(fmt.Errorf("try adding an amount to an invalid token: [%s, %s, %s] <> [%s, %s,%s]",
			a.token.Network(), a.token.Address(), a.token.Ticker(),
			a2.token.Network(), a2.token.Address(), a2.token.Ticker()))
	}
	return Amount{
		token: a.token,
		value: new(big.Int).Add(a.value, a2.value),
	}
}

func (a Amount) Sub(a2 Amount) Amount {
	if !a.token.Eq(a2.token) {
		panic(fmt.Errorf("try subtracting an amount to an invalid token: [%s, %s, %s] <> [%s, %s,%s]",
			a.token.Network(), a.token.Address(), a.token.Ticker(),
			a2.token.Network(), a2.token.Address(), a2.token.Ticker()))
	}
	return Amount{
		token: a.token,
		value: new(big.Int).Add(a.value, a2.value),
	}
}

func (a Amount) Mul(val interface{}) Amount {
	value := unknownValueToInt(a.token, val)
	return Amount{
		token: a.token,
		value: new(big.Int).Mul(a.value, value),
	}
}

func (a Amount) Div(val interface{}) Amount {
	value := unknownValueToInt(a.token, val)
	return Amount{
		token: a.token,
		value: new(big.Int).Quo(a.value, value),
	}
}

func (a Amount) HumanValue() interface{} {
	return a.token.ToHumanValue(a.value)
}

func unknownValueToInt(t *token.Token, val interface{}) *big.Int {
	value := big.NewInt(0)

	switch val.(type) {
	case *big.Int:
		value = val.(*big.Int)
		break
	case int64:
		value = new(big.Int).SetInt64(val.(int64))
		break
	case int:
		value = new(big.Int).SetInt64(int64(val.(int)))
		break
	case *big.Float:
		value = t.ToBaseValue(val.(*big.Float))
		break
	case float64:
		value = t.ToBaseValue(big.NewFloat(val.(float64)))
		break
	case Amount:
		value = val.(Amount).Value()
		break
	}

	return value
}

func NewAmount(t *token.Token, val interface{}) Amount {
	return Amount{
		token: t,
		value: unknownValueToInt(t, val),
	}
}
