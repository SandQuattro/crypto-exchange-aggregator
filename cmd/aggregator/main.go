package main

import (
	"crypto-exchange-agg/config"
	"crypto-exchange-agg/internal/application"
	"log"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app := application.NewApplication()
	err = app.Run(cfg)
	if err != nil {
		return
	}
}
