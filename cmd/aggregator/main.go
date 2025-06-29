package main

import (
	"context"

	"crypto-exchange-agg/config"
	"crypto-exchange-agg/internal/application"
	"crypto-exchange-agg/pkg/logging"
)

const CTX = "main"

func main() {
	ctx := context.Background()

	logging.InitLogger(true)

	ctx = logging.GetCtxWithScope(ctx, CTX)
	logger := logging.GetCtxLogger(ctx)

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal().AnErr("Config error: %s", err)
	}

	app := application.NewApplication()

	logger.Info().Msg("Application started")

	err = app.Run(ctx, cfg)
	if err != nil {
		logger.Fatal().AnErr("Application running error", err)
		return
	}
}
