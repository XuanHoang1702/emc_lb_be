package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"emc_lb/src/pkg/entities"

	"github.com/redis/go-redis/v9"
)

const (
	productListKeyPrefix = "cache:products:list:"
	productItemPrefix    = "cache:product:"
	productCacheTTL      = 10 * time.Minute
)

type ProductCacheStore interface {
	GetList(ctx context.Context, queryHash string) ([]entities.ProductResponse, error)
	SetList(ctx context.Context, queryHash string, products []entities.ProductResponse) error
	GetByID(ctx context.Context, id string) (entities.ProductResponse, error)
	SetByID(ctx context.Context, id string, product entities.ProductResponse) error
	Invalidate(ctx context.Context, id string) error
	InvalidateList(ctx context.Context) error
}

type RedisProductCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisProductCacheStore(client *redis.Client) ProductCacheStore {
	return &RedisProductCacheStore{client: client, ttl: productCacheTTL}
}

func (s *RedisProductCacheStore) GetList(ctx context.Context, queryHash string) ([]entities.ProductResponse, error) {
	start := time.Now()
	key := productListKeyPrefix + queryHash
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			CacheMissesTotal.WithLabelValues("product_list").Inc()
		} else {
			CacheGetErrorsTotal.WithLabelValues("product_list").Inc()
		}
		return nil, err
	}

	var products []entities.ProductResponse
	if err := json.Unmarshal(data, &products); err != nil {
		CacheGetErrorsTotal.WithLabelValues("product_list").Inc()
		// If data is corrupted, delete it
		_ = s.client.Del(ctx, key)
		return nil, err
	}

	CacheHitsTotal.WithLabelValues("product_list").Inc()
	CacheGetDuration.WithLabelValues("product_list").Observe(time.Since(start).Seconds())
	return products, nil
}

func (s *RedisProductCacheStore) SetList(ctx context.Context, queryHash string, products []entities.ProductResponse) error {
	start := time.Now()
	key := productListKeyPrefix + queryHash
	data, err := json.Marshal(products)
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("product_list").Inc()
		return err
	}

	err = s.client.Set(ctx, key, data, s.ttl).Err()
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("product_list").Inc()
		return err
	}
	CacheSetDuration.WithLabelValues("product_list").Observe(time.Since(start).Seconds())
	return nil
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

	err = s.client.Set(ctx, key, data, s.ttl).Err()
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
	return s.InvalidateList(ctx)
}

func (s *RedisProductCacheStore) InvalidateList(ctx context.Context) error {
	// For lists, we'll use a short TTL in general, but since this project
	// expects invalidation, we can use pattern-based deletion for now
	// with SCAN, which is acceptable since the dataset might not be huge.
	iter := s.client.Scan(ctx, 0, productListKeyPrefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		err := s.client.Del(ctx, keys...).Err()
		if err != nil {
			return err
		}
		CacheInvalidationsTotal.WithLabelValues("product_list").Add(float64(len(keys)))
	}
	return nil
}

func buildProductItemKey(id string) string {
	return fmt.Sprintf("%s%s", productItemPrefix, id)
}
