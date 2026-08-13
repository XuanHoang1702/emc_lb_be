package utils

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(buildPasswordInput(password)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(buildPasswordInput(password)))
}

func buildPasswordInput(password string) string {
	systemSecret := GetEnvRequired("SYSTEM_SECRET")
	sum := sha256.Sum256([]byte(password + ":" + systemSecret))
	return hex.EncodeToString(sum[:])
}
