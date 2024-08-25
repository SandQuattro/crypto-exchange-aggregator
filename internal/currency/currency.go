package currency

//go:generate stringer -type=Cryptocurrency
type Cryptocurrency int

const (
	_ Cryptocurrency = iota
	EUR
	USD
	USDT
	USDC
	BTC
	ETH
	LTC
	DOGE
)
