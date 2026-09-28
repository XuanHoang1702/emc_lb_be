package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionData struct {
	UserID         string
	SessionVersion int32
}

type RefreshTokenStore interface {
	Save(ctx context.Context, session SessionData, refreshToken string, ttl time.Duration) error
	GetSession(ctx context.Context, refreshToken string) (SessionData, error)
	Delete(ctx context.Context, refreshToken string) error
}

type RedisRefreshTokenStore struct {
	client *redis.Client
}

func NewRedisRefreshTokenStore(client *redis.Client) RefreshTokenStore {
	return &RedisRefreshTokenStore{client: client}
}

func (s *RedisRefreshTokenStore) Save(ctx context.Context, session SessionData, refreshToken string, ttl time.Duration) error {
	val := fmt.Sprintf("%s:%d", session.UserID, session.SessionVersion)
	return s.client.Set(ctx, buildRefreshTokenKey(refreshToken), val, ttl).Err()
}

func (s *RedisRefreshTokenStore) GetSession(ctx context.Context, refreshToken string) (SessionData, error) {
	val, err := s.client.Get(ctx, buildRefreshTokenKey(refreshToken)).Result()
	if err != nil {
		return SessionData{}, err
	}

	var session SessionData
	parts := strings.Split(val, ":")
	if len(parts) == 2 {
		session.UserID = parts[0]
		var version int
		_, err = fmt.Sscanf(parts[1], "%d", &version)
		if err == nil {
			session.SessionVersion = int32(version)
			return session, nil
		}
	}
	
	// Fallback for older tokens that only had userID (if any exist)
	return SessionData{UserID: val, SessionVersion: 1}, nil
}

func (s *RedisRefreshTokenStore) Delete(ctx context.Context, refreshToken string) error {
	return s.client.Del(ctx, buildRefreshTokenKey(refreshToken)).Err()
}

func buildRefreshTokenKey(refreshToken string) string {
	return fmt.Sprintf("refresh_token:%s", refreshToken)
}
