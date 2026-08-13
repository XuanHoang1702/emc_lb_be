package utils

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var loadEnvOnce sync.Once

func LoadEnv() {
	loadEnvOnce.Do(func() {
		_ = godotenv.Load()
	})
}

func GetEnv(key string, fallback string) string {
	LoadEnv()

	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func GetEnvRequired(key string) string {
	LoadEnv()
	value := os.Getenv(key)
	if value == "" {
		panic("Required environment variable " + key + " is not set")
	}
	return value
}
