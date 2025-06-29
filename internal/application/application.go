package application

import (
	"context"
	"fmt"
	"net/http"

	"crypto-exchange-agg/pkg/logging"

	"crypto-exchange-agg/config"
	"crypto-exchange-agg/internal/currency"
	"crypto-exchange-agg/internal/providers"

	"golang.org/x/sync/errgroup"
)

const CTX = "application"

type Application struct{}

func NewApplication() *Application {
	return &Application{}
}

func (a Application) Run(ctx context.Context, cfg *config.Config) error {
	ctx = logging.GetCtxWithScope(logging.GetCtxWithTraceId(ctx), CTX)
	logger := logging.GetCtxLogger(ctx)

	client := http.DefaultClient

	from := []currency.Cryptocurrency{currency.EUR, currency.USDT, currency.USDC, currency.BTC, currency.ETH, currency.LTC, currency.DOGE}
	to := []currency.Cryptocurrency{currency.EUR, currency.USD}

	coinGate := providers.CoinGate{
		Client: client,
	}

	_ = providers.CoinApi{
		Logger: logger,
		Client: client,
		Config: cfg,
	}

	g, _ := errgroup.WithContext(ctx)

	for _, currencyFrom := range from {
		for _, currencyTo := range to {
			g.Go(func() error {
				rate, err := coinGate.GetMerchantRate(currencyFrom, currencyTo)
				if err != nil {
					return err
				}

				logger.Info().Msg(fmt.Sprintf("%s_%s: %s", currencyFrom, currencyTo, rate))
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		logger.Fatal().AnErr("Application running error", err)
	}

	return nil
}
