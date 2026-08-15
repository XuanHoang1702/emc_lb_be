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
	shopOwnerPrefix = "cache:shop:owner:"
	shopCacheTTL    = 15 * time.Minute
)

type ShopCacheStore interface {
	GetByOwnerID(ctx context.Context, ownerID string) (entities.ShopResponse, error)
	SetByOwnerID(ctx context.Context, ownerID string, shop entities.ShopResponse) error
	InvalidateByOwnerID(ctx context.Context, ownerID string) error
}

type RedisShopCacheStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisShopCacheStore(client *redis.Client) ShopCacheStore {
	return &RedisShopCacheStore{client: client, ttl: shopCacheTTL}
}

func (s *RedisShopCacheStore) GetByOwnerID(ctx context.Context, ownerID string) (entities.ShopResponse, error) {
	data, err := s.client.Get(ctx, buildShopOwnerKey(ownerID)).Bytes()
	if err != nil {
		return entities.ShopResponse{}, err
	}

	var shop entities.ShopResponse
	if err := json.Unmarshal(data, &shop); err != nil {
		return entities.ShopResponse{}, err
	}

	return shop, nil
}

func (s *RedisShopCacheStore) SetByOwnerID(ctx context.Context, ownerID string, shop entities.ShopResponse) error {
	data, err := json.Marshal(shop)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, buildShopOwnerKey(ownerID), data, s.ttl).Err()
}

func (s *RedisShopCacheStore) InvalidateByOwnerID(ctx context.Context, ownerID string) error {
	return s.client.Del(ctx, buildShopOwnerKey(ownerID)).Err()
}

func buildShopOwnerKey(ownerID string) string {
	return fmt.Sprintf("%s%s", shopOwnerPrefix, ownerID)
}
