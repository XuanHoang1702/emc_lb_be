package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type EmailOTPStore interface {
	Save(ctx context.Context, email string, otp string, ttl time.Duration) error
	Get(ctx context.Context, email string) (string, error)
	Delete(ctx context.Context, email string) error
}

type RedisEmailOTPStore struct {
	client *redis.Client
}

func NewRedisEmailOTPStore(client *redis.Client) EmailOTPStore {
	return &RedisEmailOTPStore{client: client}
}

func (s *RedisEmailOTPStore) Save(ctx context.Context, email string, otp string, ttl time.Duration) error {
	return s.client.Set(ctx, buildEmailOTPKey(email), otp, ttl).Err()
}

func (s *RedisEmailOTPStore) Get(ctx context.Context, email string) (string, error) {
	return s.client.Get(ctx, buildEmailOTPKey(email)).Result()
}

func (s *RedisEmailOTPStore) Delete(ctx context.Context, email string) error {
	return s.client.Del(ctx, buildEmailOTPKey(email)).Err()
}

func buildEmailOTPKey(email string) string {
	return fmt.Sprintf("email_otp:%s", email)
}
