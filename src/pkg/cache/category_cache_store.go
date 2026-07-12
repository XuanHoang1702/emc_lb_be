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
	categoryListKey    = "cache:categories:list"
	categoryItemPrefix = "cache:category:"
	categoryCacheTTL   = 30 * time.Minute
)

type CategoryCacheStore interface {
	GetAll(ctx context.Context) ([]entities.CategoryResponse, error)
	SetAll(ctx context.Context, categories []entities.CategoryResponse) error
	GetByID(ctx context.Context, id string) (entities.CategoryResponse, error)
	SetByID(ctx context.Context, id string, category entities.CategoryResponse) error
	Invalidate(ctx context.Context, id string) error
	InvalidateAll(ctx context.Context) error
}

type RedisCategoryCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCategoryCacheStore(client *redis.Client) CategoryCacheStore {
	return &RedisCategoryCacheStore{client: client, ttl: categoryCacheTTL}
}

func (s *RedisCategoryCacheStore) GetAll(ctx context.Context) ([]entities.CategoryResponse, error) {
	data, err := s.client.Get(ctx, categoryListKey).Bytes()
	if err != nil {
		return nil, err
	}

	var categories []entities.CategoryResponse
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *RedisCategoryCacheStore) SetAll(ctx context.Context, categories []entities.CategoryResponse) error {
	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, categoryListKey, data, s.ttl).Err()
}

func (s *RedisCategoryCacheStore) GetByID(ctx context.Context, id string) (entities.CategoryResponse, error) {
	data, err := s.client.Get(ctx, buildCategoryItemKey(id)).Bytes()
	if err != nil {
		return entities.CategoryResponse{}, err
	}

	var category entities.CategoryResponse
	if err := json.Unmarshal(data, &category); err != nil {
		return entities.CategoryResponse{}, err
	}

	return category, nil
}

func (s *RedisCategoryCacheStore) SetByID(ctx context.Context, id string, category entities.CategoryResponse) error {
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, buildCategoryItemKey(id), data, s.ttl).Err()
}

func (s *RedisCategoryCacheStore) Invalidate(ctx context.Context, id string) error {
	return s.client.Del(ctx, buildCategoryItemKey(id), categoryListKey).Err()
}

func (s *RedisCategoryCacheStore) InvalidateAll(ctx context.Context) error {
	iter := s.client.Scan(ctx, 0, categoryItemPrefix+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	keys = append(keys, categoryListKey)

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

func buildCategoryItemKey(id string) string {
	return fmt.Sprintf("%s%s", categoryItemPrefix, id)
}
