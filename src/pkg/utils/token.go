package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"emc_lb/src/pkg/config"

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

func GenerateTokenPair(userID string, role string, cfg config.JWTSettings) (TokenPair, error) {
	accessToken, accessExpiresAt, err := GenerateAccessToken(userID, role, cfg)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, refreshExpiresAt, err := GenerateRefreshToken(cfg)
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

func GenerateAccessToken(userID string, role string, cfg config.JWTSettings) (string, time.Time, error) {
	now := time.Now()
	return generateAccessJWT(userID, role, now, cfg.AccessTTL, buildTokenSecret(cfg.AccessSecret))
}

func ParseAccessToken(accessToken string, accessSecret string) (string, string, error) {
	parsedToken, err := jwt.ParseWithClaims(accessToken, &accessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}

		return buildTokenSecret(accessSecret), nil
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

func GenerateRefreshToken(cfg config.JWTSettings) (string, time.Time, error) {
	expiresAt := time.Now().Add(cfg.RefreshTTL)

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, err
	}

	payload := base64.RawURLEncoding.EncodeToString(tokenBytes)
	signature := signTokenPayload(payload, buildTokenSecret(cfg.RefreshSecret))
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
	// For simplicity, we just hash the secret itself if no systemSecret is present,
	// or we can just return the secret as bytes. To match previous behavior where
	// systemSecret = AccessSecret, we can just hash it against itself.
	sum := sha256.Sum256([]byte(secret + ":" + secret))
	return sum[:]
}

func signTokenPayload(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
