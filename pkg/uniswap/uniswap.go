package uniswap

type Uniswap interface {
	FindPool()
	GetPool()
	IncreaseLiquidity()
	DecreaseLiquidity()
	GetPosition()
	CollectFees()
}
