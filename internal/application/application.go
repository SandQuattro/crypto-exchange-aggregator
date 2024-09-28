package application

import (
	"context"
	"log"
	"net/http"

	"crypto-exchange-agg/config"
	"crypto-exchange-agg/internal/currency"
	"crypto-exchange-agg/internal/providers"
	"golang.org/x/sync/errgroup"
)

type Application struct{}

func NewApplication() *Application {
	return &Application{}
}

func (a Application) Run(cfg *config.Config) error {
	client := http.DefaultClient

	from := []currency.Cryptocurrency{currency.EUR, currency.USDT, currency.USDC, currency.BTC, currency.ETH, currency.LTC, currency.DOGE}
	to := []currency.Cryptocurrency{currency.EUR, currency.USD}

	coinGate := providers.CoinGate{
		Client: client,
	}

	coinAPI := providers.CoinApi{
		Client: client,
		Config: cfg,
	}

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		currencies, err := coinAPI.GetUSDRates(from)
		if err != nil {
			return err
		}
		log.Println("[COIN API](USD RATES) ", currencies)
		return nil
	})

	g.Go(func() error {
		rates, err := coinGate.GetAllRates()
		if err != nil {
			return err
		}
		log.Println("[COIN GATE](ALL RATES) ", rates)
		return nil
	})

	for _, currencyFrom := range from {
		for _, currencyTo := range to {
			g.Go(func() error {
				rate, err := coinGate.GetMerchantRate(currencyFrom, currencyTo)
				if err != nil {
					return err
				}

				log.Printf("Currency %s rate to %s: %s", currencyFrom, currencyTo, rate)
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		log.Println(err)
	}

	return nil
}
