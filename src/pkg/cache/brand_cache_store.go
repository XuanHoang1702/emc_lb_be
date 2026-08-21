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
	brandListKey    = "cache:brands:list"
	brandItemPrefix = "cache:brand:"
	brandCacheTTL   = 10 * time.Minute
)

type BrandCacheStore interface {
	GetAll(ctx context.Context) ([]entities.BrandResponse, error)
	SetAll(ctx context.Context, brands []entities.BrandResponse) error
	GetByID(ctx context.Context, id string) (entities.BrandResponse, error)
	SetByID(ctx context.Context, id string, brand entities.BrandResponse) error
	Invalidate(ctx context.Context, id string) error
	InvalidateAll(ctx context.Context) error
}

type RedisBrandCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisBrandCacheStore(client *redis.Client) BrandCacheStore {
	return &RedisBrandCacheStore{client: client, ttl: brandCacheTTL}
}

func (s *RedisBrandCacheStore) GetAll(ctx context.Context) ([]entities.BrandResponse, error) {
	data, err := s.client.Get(ctx, brandListKey).Bytes()
	if err != nil {
		return nil, err
	}

	var brands []entities.BrandResponse
	if err := json.Unmarshal(data, &brands); err != nil {
		return nil, err
	}

	return brands, nil
}

func (s *RedisBrandCacheStore) SetAll(ctx context.Context, brands []entities.BrandResponse) error {
	data, err := json.Marshal(brands)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, brandListKey, data, s.ttl).Err()
}

func (s *RedisBrandCacheStore) GetByID(ctx context.Context, id string) (entities.BrandResponse, error) {
	data, err := s.client.Get(ctx, buildBrandItemKey(id)).Bytes()
	if err != nil {
		return entities.BrandResponse{}, err
	}

	var brand entities.BrandResponse
	if err := json.Unmarshal(data, &brand); err != nil {
		return entities.BrandResponse{}, err
	}

	return brand, nil
}

func (s *RedisBrandCacheStore) SetByID(ctx context.Context, id string, brand entities.BrandResponse) error {
	data, err := json.Marshal(brand)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, buildBrandItemKey(id), data, s.ttl).Err()
}

func (s *RedisBrandCacheStore) Invalidate(ctx context.Context, id string) error {
	return s.client.Del(ctx, buildBrandItemKey(id), brandListKey).Err()
}

func (s *RedisBrandCacheStore) InvalidateAll(ctx context.Context) error {
	iter := s.client.Scan(ctx, 0, brandItemPrefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	keys = append(keys, brandListKey)

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

func buildBrandItemKey(id string) string {
	return fmt.Sprintf("%s%s", brandItemPrefix, id)
}
