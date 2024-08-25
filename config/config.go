package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"path/filepath"
	"runtime"
)

type (
	// Config -.
	Config struct {
		App  `json:"app"`
		HTTP `json:"http"`
		Log  `json:"logger"`
		Keys `json:"keys"`
	}

	// App -.
	App struct {
		Name    string `env-required:"false" json:"name"    env:"APP_NAME"`
		Version string `env-required:"false" json:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port int `env-required:"false" json:"port" env:"HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"false" json:"level"   env:"LOG_LEVEL"`
	}

	CoinApi struct {
		Key string `env-required:"true" json:"key" env:"COIN_API_KEY"`
	}

	Keys struct {
		CoinApi `env-required:"false" json:"coin_api"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	configPath := filepath.Join(basepath, "config.json")
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
