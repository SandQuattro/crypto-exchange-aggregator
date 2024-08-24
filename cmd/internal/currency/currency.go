package currency

type Cryptocurrency int

const (
	EUR Cryptocurrency = iota
	USD
	USDT
	USDC
	BTC
	ETH
	LTC
	DOGE
)

func (d Cryptocurrency) String() string {
	return [...]string{"EUR", "USD", "USDT", "USDC", "BTC", "ETH", "LTC", "DOGE"}[d]
}
