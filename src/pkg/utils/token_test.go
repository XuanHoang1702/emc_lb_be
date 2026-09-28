package utils

import (
	"strings"
	"testing"
	"time"

	"emc_lb/src/pkg/config"
)

func TestGenerateTokenPairAndParse(t *testing.T) {
	cfg := config.JWTSettings{
		AccessSecret:  "test-access-secret-key-12345",
		RefreshSecret: "test-refresh-secret-key-12345",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
	}

	t.Run("generates valid token pair", func(t *testing.T) {
		pair, err := GenerateTokenPair("user-123", "admin", cfg)
		if err != nil {
			t.Fatalf("GenerateTokenPair: %v", err)
		}

		if pair.AccessToken == "" {
			t.Error("AccessToken is empty")
		}
		if pair.RefreshToken == "" {
			t.Error("RefreshToken is empty")
		}
		if pair.AccessTokenExpiresAt.IsZero() {
			t.Error("AccessTokenExpiresAt is zero")
		}
		if pair.RefreshTokenExpiresAt.IsZero() {
			t.Error("RefreshTokenExpiresAt is zero")
		}
	})

	t.Run("access token round-trips through parse", func(t *testing.T) {
		pair, err := GenerateTokenPair("user-456", "member", cfg)
		if err != nil {
			t.Fatalf("GenerateTokenPair: %v", err)
		}

		userID, role, err := ParseAccessToken(pair.AccessToken, cfg.AccessSecret)
		if err != nil {
			t.Fatalf("ParseAccessToken: %v", err)
		}
		if userID != "user-456" {
			t.Errorf("userID = %q; want %q", userID, "user-456")
		}
		if role != "member" {
			t.Errorf("role = %q; want %q", role, "member")
		}
	})

	t.Run("parse rejects wrong secret", func(t *testing.T) {
		pair, err := GenerateTokenPair("user-789", "admin", cfg)
		if err != nil {
			t.Fatalf("GenerateTokenPair: %v", err)
		}

		_, _, err = ParseAccessToken(pair.AccessToken, "wrong-secret")
		if err == nil {
			t.Error("ParseAccessToken with wrong secret should fail")
		}
	})

	t.Run("parse rejects garbage token", func(t *testing.T) {
		_, _, err := ParseAccessToken("not-a-valid-jwt", cfg.AccessSecret)
		if err == nil {
			t.Error("ParseAccessToken with garbage input should fail")
		}
	})

	t.Run("parse rejects empty token", func(t *testing.T) {
		_, _, err := ParseAccessToken("", cfg.AccessSecret)
		if err == nil {
			t.Error("ParseAccessToken with empty token should fail")
		}
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	cfg := config.JWTSettings{
		RefreshSecret: "test-refresh-secret",
		RefreshTTL:    24 * time.Hour,
	}

	t.Run("produces payload.signature format", func(t *testing.T) {
		token, _, err := GenerateRefreshToken(cfg)
		if err != nil {
			t.Fatalf("GenerateRefreshToken: %v", err)
		}

		parts := strings.SplitN(token, ".", 2)
		if len(parts) != 2 {
			t.Fatalf("refresh token should have 2 parts separated by '.', got %d", len(parts))
		}
		if parts[0] == "" {
			t.Error("payload part is empty")
		}
		if parts[1] == "" {
			t.Error("signature part is empty")
		}
	})

	t.Run("signature is full HMAC-SHA256 (64 hex chars)", func(t *testing.T) {
		token, _, err := GenerateRefreshToken(cfg)
		if err != nil {
			t.Fatalf("GenerateRefreshToken: %v", err)
		}

		parts := strings.SplitN(token, ".", 2)
		sig := parts[1]
		if len(sig) != 64 {
			t.Errorf("signature length = %d; want 64 (full HMAC-SHA256 hex)", len(sig))
		}
	})

	t.Run("each token is unique", func(t *testing.T) {
		token1, _, _ := GenerateRefreshToken(cfg)
		token2, _, _ := GenerateRefreshToken(cfg)
		if token1 == token2 {
			t.Error("two successive refresh tokens should be different")
		}
	})

	t.Run("expiry is in the future", func(t *testing.T) {
		before := time.Now()
		_, expiresAt, err := GenerateRefreshToken(cfg)
		if err != nil {
			t.Fatalf("GenerateRefreshToken: %v", err)
		}
		if !expiresAt.After(before) {
			t.Error("expiresAt should be after now")
		}
	})
}

func TestSignTokenPayloadDeterministic(t *testing.T) {
	secret := buildTokenSecret("my-secret")
	payload := "test-payload-data"

	sig1 := signTokenPayload(payload, secret)
	sig2 := signTokenPayload(payload, secret)

	if sig1 != sig2 {
		t.Error("same payload+secret should produce identical signatures")
	}
	if len(sig1) != 64 {
		t.Errorf("signature length = %d; want 64", len(sig1))
	}
}

func TestSignTokenPayloadDifferentInputs(t *testing.T) {
	secret := buildTokenSecret("my-secret")

	sig1 := signTokenPayload("payload-a", secret)
	sig2 := signTokenPayload("payload-b", secret)

	if sig1 == sig2 {
		t.Error("different payloads should produce different signatures")
	}
}

func TestBuildTokenSecretDeterministic(t *testing.T) {
	s1 := buildTokenSecret("secret-key")
	s2 := buildTokenSecret("secret-key")

	if len(s1) != 32 {
		t.Errorf("secret length = %d; want 32 (SHA-256)", len(s1))
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			t.Fatal("same input should produce identical secret")
		}
	}
}

func TestBuildTokenSecretDifferentInputs(t *testing.T) {
	s1 := buildTokenSecret("key-a")
	s2 := buildTokenSecret("key-b")

	same := true
	for i := range s1 {
		if s1[i] != s2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("different inputs should produce different secrets")
	}
}
