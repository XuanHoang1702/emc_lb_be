package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type accessTokenClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

func GenerateTokenPair(userID string, role string) (TokenPair, error) {
	accessToken, accessExpiresAt, err := GenerateAccessToken(userID, role)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, refreshExpiresAt, err := GenerateRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

func GenerateAccessToken(userID string, role string) (string, time.Time, error) {
	ttl := GetDurationFromEnv("ACCESS_TOKEN_TTL", 15*time.Minute)
	now := time.Now()
	return generateAccessJWT(userID, role, now, ttl, buildTokenSecret(GetEnv("ACCESS_TOKEN_SECRET", "access-secret")))
}

func ParseAccessToken(accessToken string) (string, string, error) {
	parsedToken, err := jwt.ParseWithClaims(accessToken, &accessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}

		return buildTokenSecret(GetEnv("ACCESS_TOKEN_SECRET", "access-secret")), nil
	})
	if err != nil {
		return "", "", err
	}

	claims, ok := parsedToken.Claims.(*accessTokenClaims)
	if !ok || !parsedToken.Valid || claims.UserID == "" {
		return "", "", errors.New("invalid access token")
	}

	return claims.UserID, claims.Role, nil
}

func GenerateRefreshToken() (string, time.Time, error) {
	ttl := GetDurationFromEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	expiresAt := time.Now().Add(ttl)

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, err
	}

	payload := base64.RawURLEncoding.EncodeToString(tokenBytes)
	signature := signTokenPayload(payload, buildTokenSecret(GetEnv("REFRESH_TOKEN_SECRET", "refresh-secret")))
	refreshToken := payload + "." + signature
	return refreshToken, expiresAt, nil
}

func generateAccessJWT(userID string, role string, issuedAt time.Time, ttl time.Duration, secret []byte) (string, time.Time, error) {
	expiresAt := issuedAt.Add(ttl)
	claims := accessTokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

func buildTokenSecret(secret string) []byte {
	systemSecret := GetEnv("SYSTEM_SECRET", "system-secret")
	sum := sha256.Sum256([]byte(systemSecret + ":" + secret))
	return sum[:]
}

func signTokenPayload(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))
	return strings.ToLower(signature[:24])
}
