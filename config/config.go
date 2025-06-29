package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App  `json:"app"`
		HTTP `json:"http"`
		Log  `json:"logger"`
		Keys `json:"keys"`
	}

	App struct {
		Name     string `env-required:"false" json:"name"     env:"APP_NAME"`
		Version  string `env-required:"false" json:"version"  env:"APP_VERSION"`
		Provider string `env-required:"false" json:"provider" env:"CRYPTO_PROVIDER" env-default:"coingate"`
	}

	HTTP struct {
		Port int `env-required:"false" json:"port" env:"HTTP_PORT"`
	}

	Log struct {
		Level string `env-required:"false" json:"level"   env:"LOG_LEVEL"`
	}

	Keys struct {
		CoinMarketCap `env-required:"false" json:"coinmarketcap"`
	}

	CoinMarketCap struct {
		Key string `env-required:"false" json:"key" env:"COINMARKETCAP_API_KEY"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	configPath := filepath.Join(basePath, "config.json")

	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("env config error: %w", err)
	}

	return cfg, nil
}
