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
	productListKey    = "cache:products:list"
	productItemPrefix = "cache:product:"
	productCacheTTL   = 10 * time.Minute
)

type ProductCacheStore interface {
	GetAll(ctx context.Context) ([]entities.ProductResponse, error)
	SetAll(ctx context.Context, products []entities.ProductResponse) error
	GetByID(ctx context.Context, id string) (entities.ProductResponse, error)
	SetByID(ctx context.Context, id string, product entities.ProductResponse) error
	Invalidate(ctx context.Context, id string) error
	InvalidateAll(ctx context.Context) error
}

type RedisProductCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisProductCacheStore(client *redis.Client) ProductCacheStore {
	return &RedisProductCacheStore{client: client, ttl: productCacheTTL}
}

func (s *RedisProductCacheStore) GetAll(ctx context.Context) ([]entities.ProductResponse, error) {
	data, err := s.client.Get(ctx, productListKey).Bytes()
	if err != nil {
		return nil, err
	}

	var products []entities.ProductResponse
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *RedisProductCacheStore) SetAll(ctx context.Context, products []entities.ProductResponse) error {
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, productListKey, data, s.ttl).Err()
}

func (s *RedisProductCacheStore) GetByID(ctx context.Context, id string) (entities.ProductResponse, error) {
	data, err := s.client.Get(ctx, buildProductItemKey(id)).Bytes()
	if err != nil {
		return entities.ProductResponse{}, err
	}

	var product entities.ProductResponse
	if err := json.Unmarshal(data, &product); err != nil {
		return entities.ProductResponse{}, err
	}

	return product, nil
}

func (s *RedisProductCacheStore) SetByID(ctx context.Context, id string, product entities.ProductResponse) error {
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, buildProductItemKey(id), data, s.ttl).Err()
}

func (s *RedisProductCacheStore) Invalidate(ctx context.Context, id string) error {
	return s.client.Del(ctx, buildProductItemKey(id), productListKey).Err()
}

func (s *RedisProductCacheStore) InvalidateAll(ctx context.Context) error {
	iter := s.client.Scan(ctx, 0, productItemPrefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	keys = append(keys, productListKey)

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

func buildProductItemKey(id string) string {
	return fmt.Sprintf("%s%s", productItemPrefix, id)
}
