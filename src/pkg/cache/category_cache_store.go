package cache

import (
	"context"
	"encoding/json"
	"time"

	"emc_lb/src/pkg/entities"

	"github.com/redis/go-redis/v9"
)

const categoryListKey = "cache:category:list"

type CategoryCacheStore interface {
	GetList(ctx context.Context) ([]entities.CategoryResponse, error)
	SetList(ctx context.Context, categories []entities.CategoryResponse) error
	InvalidateList(ctx context.Context) error
}

type RedisCategoryCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCategoryCacheStore(client *redis.Client, ttl time.Duration) CategoryCacheStore {
	return &RedisCategoryCacheStore{client: client, ttl: ttl}
}

func (s *RedisCategoryCacheStore) GetList(ctx context.Context) ([]entities.CategoryResponse, error) {
	start := time.Now()
	data, err := s.client.Get(ctx, categoryListKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			CacheMissesTotal.WithLabelValues("category").Inc()
		} else {
			CacheGetErrorsTotal.WithLabelValues("category").Inc()
		}
		return nil, err
	}

	var categories []entities.CategoryResponse
	if err := json.Unmarshal(data, &categories); err != nil {
		CacheGetErrorsTotal.WithLabelValues("category").Inc()
		_ = s.client.Del(ctx, categoryListKey)
		return nil, err
	}

	CacheHitsTotal.WithLabelValues("category").Inc()
	CacheGetDuration.WithLabelValues("category").Observe(time.Since(start).Seconds())
	return categories, nil
}

func (s *RedisCategoryCacheStore) SetList(ctx context.Context, categories []entities.CategoryResponse) error {
	start := time.Now()
	data, err := json.Marshal(categories)
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("category").Inc()
		return err
	}

	err = s.client.Set(ctx, categoryListKey, data, s.ttl).Err()
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("category").Inc()
		return err
	}
	CacheSetDuration.WithLabelValues("category").Observe(time.Since(start).Seconds())
	return nil
}

func (s *RedisCategoryCacheStore) InvalidateList(ctx context.Context) error {
	err := s.client.Del(ctx, categoryListKey).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	CacheInvalidationsTotal.WithLabelValues("category").Inc()
	return nil
}
