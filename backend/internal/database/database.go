package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds connection pool tuning parameters.
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// Connect creates a configured pgxpool with production-grade pool settings.
func Connect(ctx context.Context, databaseUrl string, poolCfg *PoolConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Apply pool tuning
	if poolCfg != nil {
		if poolCfg.MaxConns > 0 {
			config.MaxConns = poolCfg.MaxConns
		}
		if poolCfg.MinConns > 0 {
			config.MinConns = poolCfg.MinConns
		}
		if poolCfg.MaxConnLifetime > 0 {
			config.MaxConnLifetime = poolCfg.MaxConnLifetime
		}
		if poolCfg.MaxConnIdleTime > 0 {
			config.MaxConnIdleTime = poolCfg.MaxConnIdleTime
		}
		if poolCfg.HealthCheckPeriod > 0 {
			config.HealthCheckPeriod = poolCfg.HealthCheckPeriod
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
