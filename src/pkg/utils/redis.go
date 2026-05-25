package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", GetEnv("REDIS_HOST", "localhost"), GetEnv("REDIS_PORT", "6379")),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	pingContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(pingContext).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
