package utils

import (
	"context"
	"fmt"
	"time"

	"emc_lb/src/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgPoolFromConfig creates a pgxpool using typed AppConfig.
// Pool settings are derived from config (MaxConns, MinConns).
func NewPgPoolFromConfig(ctx context.Context, cfg *config.PostgresSettings) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = 1 * time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

// NewPgPool creates a pgxpool using env vars directly (legacy fallback).
// Prefer NewPgPoolFromConfig when AppConfig is available.
func NewPgPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg := &config.PostgresSettings{
		URL:      GetEnv("DATABASE_URL", ""),
		Host:     GetEnv("POSTGRES_HOST", "localhost"),
		Port:     GetEnv("POSTGRES_PORT", "5432"),
		User:     GetEnv("POSTGRES_USER", "postgres"),
		Password: GetEnv("POSTGRES_PASSWORD", "postgres"),
		DB:       GetEnv("POSTGRES_DB", "emc_lb"),
		Timezone: GetEnv("POSTGRES_TIMEZONE", "Asia/Ho_Chi_Minh"),
		MaxConns: 10,
		MinConns: 2,
	}
	return NewPgPoolFromConfig(ctx, cfg)
}
