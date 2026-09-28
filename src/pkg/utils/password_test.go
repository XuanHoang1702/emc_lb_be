package utils

import (
	"testing"
)

func TestHashPasswordAndCheck(t *testing.T) {
	tests := []struct {
		name         string
		password     string
		systemSecret string
	}{
		{"simple password", "password123", "system-secret"},
		{"unicode password", "mật_khẩu_bí_mật", "secret-key"},
		{"long password", "a-very-long-password-that-exceeds-normal-length-requirements-0123456789", "sys"},
		{"empty secret", "password", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := HashPassword(tt.password, tt.systemSecret)
			if err != nil {
				t.Fatalf("HashPassword: %v", err)
			}
			if hashed == "" {
				t.Fatal("hashed password is empty")
			}
			if hashed == tt.password {
				t.Fatal("hashed password should not equal plaintext")
			}

			// Correct password should pass.
			if err := CheckPassword(tt.password, hashed, tt.systemSecret); err != nil {
				t.Errorf("CheckPassword with correct password: %v", err)
			}
		})
	}
}

func TestCheckPasswordRejectsWrongPassword(t *testing.T) {
	hashed, err := HashPassword("correct", "secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if err := CheckPassword("wrong", hashed, "secret"); err == nil {
		t.Error("CheckPassword should reject wrong password")
	}
}

func TestCheckPasswordRejectsWrongSecret(t *testing.T) {
	hashed, err := HashPassword("password", "secret-a")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if err := CheckPassword("password", hashed, "secret-b"); err == nil {
		t.Error("CheckPassword should reject wrong system secret")
	}
}

func TestHashPasswordProducesDifferentHashes(t *testing.T) {
	h1, _ := HashPassword("password", "secret")
	h2, _ := HashPassword("password", "secret")

	// bcrypt includes a random salt, so two hashes of the same input differ.
	if h1 == h2 {
		t.Error("bcrypt should produce different hashes each time (random salt)")
	}
}
