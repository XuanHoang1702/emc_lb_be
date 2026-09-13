package utils

import (
	"context"
	"fmt"
	"time"

	"emc_lb/src/pkg/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClientFromConfig creates a Redis client using typed AppConfig.
func NewRedisClientFromConfig(cfg *config.RedisSettings) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr(),
		Password:        cfg.Password,
		DB:              0,
		PoolSize:        20,
		MinIdleConns:    5,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis %s: %w", cfg.Addr(), err)
	}

	return client, nil
}

// NewRedisClient creates a Redis client using env vars directly (legacy fallback).
// Prefer NewRedisClientFromConfig when AppConfig is available.
func NewRedisClient() (*redis.Client, error) {
	cfg := &config.RedisSettings{
		Host:     GetEnv("REDIS_HOST", "localhost"),
		Port:     GetEnv("REDIS_PORT", "6379"),
		Password: GetEnv("REDIS_PASSWORD", ""),
	}
	return NewRedisClientFromConfig(cfg)
}
