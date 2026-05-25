package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshTokenStore interface {
	Save(ctx context.Context, userID string, refreshToken string, ttl time.Duration) error
	GetUserID(ctx context.Context, refreshToken string) (string, error)
	Delete(ctx context.Context, refreshToken string) error
}

type RedisRefreshTokenStore struct {
	client *redis.Client
}

func NewRedisRefreshTokenStore(client *redis.Client) RefreshTokenStore {
	return &RedisRefreshTokenStore{client: client}
}

func (s *RedisRefreshTokenStore) Save(ctx context.Context, userID string, refreshToken string, ttl time.Duration) error {
	return s.client.Set(ctx, buildRefreshTokenKey(refreshToken), userID, ttl).Err()
}

func (s *RedisRefreshTokenStore) GetUserID(ctx context.Context, refreshToken string) (string, error) {
	return s.client.Get(ctx, buildRefreshTokenKey(refreshToken)).Result()
}

func (s *RedisRefreshTokenStore) Delete(ctx context.Context, refreshToken string) error {
	return s.client.Del(ctx, buildRefreshTokenKey(refreshToken)).Err()
}

func buildRefreshTokenKey(refreshToken string) string {
	return fmt.Sprintf("refresh_token:%s", refreshToken)
}
