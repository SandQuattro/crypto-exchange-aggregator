package main

import (
	"context"
	"crypto-exchange-agg/internal/currency"
	"crypto-exchange-agg/internal/providers"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
)

func main() {
	from := []currency.Cryptocurrency{currency.EUR, currency.USD, currency.USDT, currency.USDC, currency.BTC, currency.ETH, currency.LTC, currency.DOGE}

	log.Println(from[0].String())

	coinGate := coingate.CoinGate{
		Client: http.DefaultClient,
	}

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		currencies, err := coinGate.GetAllCurrencies()
		if err != nil {
			return err
		}
		log.Println("[ALL CURRENCIES] ", currencies)
		return nil
	})

	g.Go(func() error {
		rates, err := coinGate.GetAllRates()
		if err != nil {
			return err
		}
		log.Println("[ALL RATES] ", rates)
		return nil
	})

	g.Go(func() error {
		rates, err := coinGate.GetAllTraderRates()
		if err != nil {
			return err
		}
		log.Println("[ALL TRADER RATES] ", rates)
		return nil
	})

	g.Go(func() error {
		rates, err := coinGate.GetAllMerchantRates()
		if err != nil {
			return err
		}
		log.Println("[ALL MERCHANT RATES] ", rates)
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
