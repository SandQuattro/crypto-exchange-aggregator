package main

import (
	"context"
	"crypto-exchange-agg/cmd/internal/currency"
	coingate "crypto-exchange-agg/cmd/internal/providers"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
)

func main() {
	from := []currency.Cryptocurrency{currency.EUR, currency.USD, currency.USDT, currency.USDC, currency.BTC, currency.ETH, currency.LTC, currency.DOGE}
	// to := []currency.Cryptocurrency{}

	coinGate := coingate.CoinGate{
		Client: http.DefaultClient,
	}

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		currencies, err := coinGate.GetAllCurrencies()
		if err != nil {
			return err
		}
		log.Println(currencies)
		return nil
	})

	for _, currencyFrom := range from {
		for _, currencyTo := range from {
			g.Go(func() error {
				rate, err := coinGate.GetMerchantRate(currencyFrom, currencyTo)
				if err != nil {
					return err
				}

				log.Printf("Currency %s rate to %s: %s", currencyFrom, currencyTo, rate)
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			log.Println(err)
		}

	}

}
