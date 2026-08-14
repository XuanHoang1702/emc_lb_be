package utils

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password, systemSecret string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(buildPasswordInput(password, systemSecret)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func CheckPassword(password, hashedPassword, systemSecret string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(buildPasswordInput(password, systemSecret)))
}

func buildPasswordInput(password, systemSecret string) string {
	sum := sha256.Sum256([]byte(password + ":" + systemSecret))
	return hex.EncodeToString(sum[:])
}
