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
	DeleteAllForUser(ctx context.Context, userID string) error
}

type RedisRefreshTokenStore struct {
	client *redis.Client
}

func NewRedisRefreshTokenStore(client *redis.Client) RefreshTokenStore {
	return &RedisRefreshTokenStore{client: client}
}

func (s *RedisRefreshTokenStore) Save(ctx context.Context, userID string, refreshToken string, ttl time.Duration) error {
	pipe := s.client.Pipeline()
	pipe.Set(ctx, buildRefreshTokenKey(refreshToken), userID, ttl)
	pipe.SAdd(ctx, buildUserTokensKey(userID), refreshToken)
	// Optionally set TTL on the set, but it will be refreshed on every new login.
	pipe.Expire(ctx, buildUserTokensKey(userID), ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisRefreshTokenStore) GetUserID(ctx context.Context, refreshToken string) (string, error) {
	return s.client.Get(ctx, buildRefreshTokenKey(refreshToken)).Result()
}

func (s *RedisRefreshTokenStore) Delete(ctx context.Context, refreshToken string) error {
	userID, err := s.GetUserID(ctx, refreshToken)
	
	pipe := s.client.Pipeline()
	pipe.Del(ctx, buildRefreshTokenKey(refreshToken))
	if err == nil && userID != "" {
		pipe.SRem(ctx, buildUserTokensKey(userID), refreshToken)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisRefreshTokenStore) DeleteAllForUser(ctx context.Context, userID string) error {
	userTokensKey := buildUserTokensKey(userID)
	tokens, err := s.client.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, token := range tokens {
		pipe.Del(ctx, buildRefreshTokenKey(token))
	}
	pipe.Del(ctx, userTokensKey)
	
	_, err = pipe.Exec(ctx)
	return err
}

func buildRefreshTokenKey(refreshToken string) string {
	return fmt.Sprintf("refresh_token:%s", refreshToken)
}

func buildUserTokensKey(userID string) string {
	return fmt.Sprintf("user_refresh_tokens:%s", userID)
}
