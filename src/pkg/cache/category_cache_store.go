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
	categoryListKeyPrefix = "cache:categories:list:"
	categoryItemPrefix    = "cache:category:"
	categoryCacheTTL      = 30 * time.Minute
)

type CategoryCacheStore interface {
	GetList(ctx context.Context, queryHash string) ([]entities.CategoryResponse, error)
	SetList(ctx context.Context, queryHash string, categories []entities.CategoryResponse) error
	GetByID(ctx context.Context, id string) (entities.CategoryResponse, error)
	SetByID(ctx context.Context, id string, category entities.CategoryResponse) error
	Invalidate(ctx context.Context, id string) error
	InvalidateList(ctx context.Context) error
}

type RedisCategoryCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCategoryCacheStore(client *redis.Client) CategoryCacheStore {
	return &RedisCategoryCacheStore{client: client, ttl: categoryCacheTTL}
}

func (s *RedisCategoryCacheStore) GetList(ctx context.Context, queryHash string) ([]entities.CategoryResponse, error) {
	start := time.Now()
	key := categoryListKeyPrefix + queryHash
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			CacheMissesTotal.WithLabelValues("category_list").Inc()
		} else {
			CacheGetErrorsTotal.WithLabelValues("category_list").Inc()
		}
		return nil, err
	}

	var categories []entities.CategoryResponse
	if err := json.Unmarshal(data, &categories); err != nil {
		CacheGetErrorsTotal.WithLabelValues("category_list").Inc()
		_ = s.client.Del(ctx, key)
		return nil, err
	}

	CacheHitsTotal.WithLabelValues("category_list").Inc()
	CacheGetDuration.WithLabelValues("category_list").Observe(time.Since(start).Seconds())
	return categories, nil
}

func (s *RedisCategoryCacheStore) SetList(ctx context.Context, queryHash string, categories []entities.CategoryResponse) error {
	start := time.Now()
	key := categoryListKeyPrefix + queryHash
	data, err := json.Marshal(categories)
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("category_list").Inc()
		return err
	}

	err = s.client.Set(ctx, key, data, s.ttl).Err()
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("category_list").Inc()
		return err
	}
	CacheSetDuration.WithLabelValues("category_list").Observe(time.Since(start).Seconds())
	return nil
}

func (s *RedisCategoryCacheStore) GetByID(ctx context.Context, id string) (entities.CategoryResponse, error) {
	start := time.Now()
	key := buildCategoryItemKey(id)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			CacheMissesTotal.WithLabelValues("category_detail").Inc()
		} else {
			CacheGetErrorsTotal.WithLabelValues("category_detail").Inc()
		}
		return entities.CategoryResponse{}, err
	}

	var category entities.CategoryResponse
	if err := json.Unmarshal(data, &category); err != nil {
		CacheGetErrorsTotal.WithLabelValues("category_detail").Inc()
		_ = s.client.Del(ctx, key)
		return entities.CategoryResponse{}, err
	}

	CacheHitsTotal.WithLabelValues("category_detail").Inc()
	CacheGetDuration.WithLabelValues("category_detail").Observe(time.Since(start).Seconds())
	return category, nil
}

func (s *RedisCategoryCacheStore) SetByID(ctx context.Context, id string, category entities.CategoryResponse) error {
	start := time.Now()
	key := buildCategoryItemKey(id)
	data, err := json.Marshal(category)
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("category_detail").Inc()
		return err
	}

	err = s.client.Set(ctx, key, data, s.ttl).Err()
	if err != nil {
		CacheSetErrorsTotal.WithLabelValues("category_detail").Inc()
		return err
	}
	CacheSetDuration.WithLabelValues("category_detail").Observe(time.Since(start).Seconds())
	return nil
}

func (s *RedisCategoryCacheStore) Invalidate(ctx context.Context, id string) error {
	key := buildCategoryItemKey(id)
	err := s.client.Del(ctx, key).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	CacheInvalidationsTotal.WithLabelValues("category_detail").Inc()
	return s.InvalidateList(ctx)
}

func (s *RedisCategoryCacheStore) InvalidateList(ctx context.Context) error {
	iter := s.client.Scan(ctx, 0, categoryListKeyPrefix+"*", 100).Iterator()
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
		CacheInvalidationsTotal.WithLabelValues("category_list").Add(float64(len(keys)))
	}
	return nil
}

func buildCategoryItemKey(id string) string {
	return fmt.Sprintf("%s%s", categoryItemPrefix, id)
}
