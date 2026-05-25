package utils

import (
	"crypto/rand"
	"math/big"
	"time"
)

func GenerateOTP(length int) (string, error) {
	digits := make([]byte, length)
	for index := 0; index < length; index++ {
		number, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}

		digits[index] = byte('0' + number.Int64())
	}

	return string(digits), nil
}

func GetDurationFromEnv(key string, fallback time.Duration) time.Duration {
	value := GetEnv(key, "")
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
