package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseUrl string
	Port        string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseUrl: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.DatabaseUrl == "" {
		return nil, fmt.Errorf("Database URL is not set")
	}

	return cfg, nil
}
