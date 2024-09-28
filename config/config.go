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
		Name    string `env-required:"false" json:"name"    env:"APP_NAME"`
		Version string `env-required:"false" json:"version" env:"APP_VERSION"`
	}

	HTTP struct {
		Port int `env-required:"false" json:"port" env:"HTTP_PORT"`
	}

	Log struct {
		Level string `env-required:"false" json:"level"   env:"LOG_LEVEL"`
	}

	CoinAPI struct {
		Key string `env-required:"true" json:"key" env:"COIN_API_KEY"`
	}

	Keys struct {
		CoinAPI `env-required:"false" json:"coin_api"`
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
		return nil, err
	}

	return cfg, nil
}
