package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"emc_lb/src/pkg/entities"

	"github.com/redis/go-redis/v9"
)

const (
	productItemPrefix = "cache:product:detail:"
)

type ProductCacheStore interface {
	GetByID(ctx context.Context, id string) (entities.ProductResponse, error)
	SetByID(ctx context.Context, id string, product entities.ProductResponse) error
	Invalidate(ctx context.Context, id string) error
}

type RedisProductCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisProductCacheStore(client *redis.Client, ttl time.Duration) ProductCacheStore {
	return &RedisProductCacheStore{client: client, ttl: ttl}
}

func (s *RedisProductCacheStore) GetByID(ctx context.Context, id string) (entities.ProductResponse, error) {
	start := time.Now()
	key := buildProductItemKey(id)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			CacheMissesTotal.WithLabelValues("product_detail").Inc()
		} else {
			CacheGetErrorsTotal.WithLabelValues("product_detail").Inc()
		}
		return entities.ProductResponse{}, err
	}

	var product entities.ProductResponse
	if err := json.Unmarshal(data, &product); err != nil {
		CacheGetErrorsTotal.WithLabelValues("product_detail").Inc()
		_ = s.client.Del(ctx, key)
		return entities.ProductResponse{}, err
	}

	CacheHitsTotal.WithLabelValues("product_detail").Inc()
	CacheGetDuration.WithLabelValues("product_detail").Observe(time.Since(start).Seconds())
	return product, nil
}

func (s *RedisProductCacheStore) SetByID(ctx context.Context, id string, product entities.ProductResponse) error {
	start := time.Now()
	key := buildProductItemKey(id)
	data, err := json.Marshal(product)
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("product_detail").Inc()
		return err
	}

	jitter := time.Duration(rand.Int63n(int64(60 * time.Second)))
	finalTTL := s.ttl + jitter

	err = s.client.Set(ctx, key, data, finalTTL).Err()
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("product_detail").Inc()
		return err
	}
	CacheSetDuration.WithLabelValues("product_detail").Observe(time.Since(start).Seconds())
	return nil
}

func (s *RedisProductCacheStore) Invalidate(ctx context.Context, id string) error {
	key := buildProductItemKey(id)
	err := s.client.Del(ctx, key).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	CacheInvalidationsTotal.WithLabelValues("product_detail").Inc()
	return nil
}

func buildProductItemKey(id string) string {
	return fmt.Sprintf("%s%s", productItemPrefix, id)
}
