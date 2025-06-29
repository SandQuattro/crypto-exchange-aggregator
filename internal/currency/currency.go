package currency

//go:generate go tool stringer -type=Cryptocurrency
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
