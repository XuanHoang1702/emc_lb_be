package cache

import (
	"context"
	"encoding/json"
	"time"

	"emc_lb/src/pkg/entities"

	"github.com/redis/go-redis/v9"
)

const brandListKey = "cache:brand:list"

type BrandCacheStore interface {
	GetAll(ctx context.Context) ([]entities.BrandResponse, error)
	SetAll(ctx context.Context, brands []entities.BrandResponse) error
	InvalidateAll(ctx context.Context) error
}

type RedisBrandCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisBrandCacheStore(client *redis.Client, ttl time.Duration) BrandCacheStore {
	return &RedisBrandCacheStore{client: client, ttl: ttl}
}

func (s *RedisBrandCacheStore) GetAll(ctx context.Context) ([]entities.BrandResponse, error) {
	start := time.Now()
	data, err := s.client.Get(ctx, brandListKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			CacheMissesTotal.WithLabelValues("brand").Inc()
		} else {
			CacheGetErrorsTotal.WithLabelValues("brand").Inc()
		}
		return nil, err
	}

	var brands []entities.BrandResponse
	if err := json.Unmarshal(data, &brands); err != nil {
		CacheGetErrorsTotal.WithLabelValues("brand").Inc()
		_ = s.client.Del(ctx, brandListKey)
		return nil, err
	}

	CacheHitsTotal.WithLabelValues("brand").Inc()
	CacheGetDuration.WithLabelValues("brand").Observe(time.Since(start).Seconds())
	return brands, nil
}

func (s *RedisBrandCacheStore) SetAll(ctx context.Context, brands []entities.BrandResponse) error {
	start := time.Now()
	data, err := json.Marshal(brands)
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("brand").Inc()
		return err
	}

	err = s.client.Set(ctx, brandListKey, data, s.ttl).Err()
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("brand").Inc()
		return err
	}
	CacheSetDuration.WithLabelValues("brand").Observe(time.Since(start).Seconds())
	return nil
}

func (s *RedisBrandCacheStore) InvalidateAll(ctx context.Context) error {
	err := s.client.Del(ctx, brandListKey).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	CacheInvalidationsTotal.WithLabelValues("brand").Inc()
	return nil
}
