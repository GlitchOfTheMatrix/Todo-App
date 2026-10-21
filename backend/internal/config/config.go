package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseUrl string
	Port        string
	JWTSecret   string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseUrl: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	if cfg.DatabaseUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	if cfg.Port == "" {
		cfg.Port = "8000"
	}

	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "super-secret-dev-key-change-in-production"
	}

	return cfg, nil
}
