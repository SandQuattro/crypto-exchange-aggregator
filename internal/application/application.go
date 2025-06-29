package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

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
	if cfg == nil {
		return errors.New("config is nil")
	}

	ctx = logging.GetCtxWithScope(logging.GetCtxWithTraceId(ctx), CTX)
	logger := logging.GetCtxLogger(ctx)

	// Инициализируем CoinMarketCap провайдер
	client := &providers.CoinMarketCap{
		Logger: logger,
		Client: http.DefaultClient,
		Config: cfg,
	}

	logger.Info().
		Str("provider", cfg.App.Provider).
		Msg("Using crypto provider")

	symbols := []currency.Cryptocurrency{
		currency.USDT, currency.USDC,
		currency.BTC, currency.ETH, currency.LTC, currency.DOGE,
	}
	targetCurrencies := []string{"USD", "EUR"}

	// Создаем ticker для периодического вызова каждые 1 минуту и 5 секунд
	interval := time.Minute + 5*time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info().
		Str("interval", interval.String()).
		Msg("Starting periodic rate fetching")

	// Вызываем сразу при старте
	if err := a.fetchRates(ctx, client, symbols, targetCurrencies, cfg.App.Provider); err != nil {
		logger.Error().Err(err).Msg("Initial rate fetch failed")
	}

	// Периодические вызовы
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Application context cancelled, stopping...")
			return ctx.Err()
		case <-ticker.C:
			logger.Info().Msg("Ticker triggered, fetching rates...")
			if err := a.fetchRates(ctx, client, symbols, targetCurrencies, cfg.App.Provider); err != nil {
				logger.Error().Err(err).Msg("Periodic rate fetch failed")
			}
		}
	}
}

// fetchRates выполняет batch запросы для получения курсов валют
func (a Application) fetchRates(
	ctx context.Context,
	client *providers.CoinMarketCap,
	symbols []currency.Cryptocurrency,
	targetCurrencies []string,
	provider string,
) error {
	logger := logging.GetCtxLogger(ctx)

	g, gCtx := errgroup.WithContext(ctx)

	// Делаем только 2 batch запроса вместо 14 отдельных
	for _, targetCurrency := range targetCurrencies {
		target := targetCurrency

		g.Go(func() error {
			logger.Info().
				Str("target_currency", target).
				Int("symbols_count", len(symbols)).
				Msg("Fetching rates batch")

			rates, rateErr := client.GetRates(gCtx, symbols, target)
			if rateErr != nil {
				logger.Error().
					Err(rateErr).
					Str("target_currency", target).
					Str("provider", provider).
					Msg("Failed to get batch rates")
				return fmt.Errorf("failed to get batch rates for %s: %w", target, rateErr)
			}

			// Логируем успешные результаты
			for symbol, rate := range rates {
				logger.Info().
					Str("from", symbol).
					Str("to", target).
					Str("rate", rate).
					Str("provider", provider).
					Msg("Rate retrieved successfully")
			}

			logger.Info().
				Str("target_currency", target).
				Int("rates_received", len(rates)).
				Msg("Batch rates completed")

			return nil
		})
	}

	if waitErr := g.Wait(); waitErr != nil {
		return waitErr
	}

	logger.Info().Msg("All rate batches completed successfully")
	return nil
}
