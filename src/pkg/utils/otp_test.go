package utils

import (
	"os"
	"testing"
	"time"
)

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"4-digit OTP", 4},
		{"6-digit OTP", 6},
		{"8-digit OTP", 8},
		{"1-digit OTP", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otp, err := GenerateOTP(tt.length)
			if err != nil {
				t.Fatalf("GenerateOTP(%d): %v", tt.length, err)
			}

			if len(otp) != tt.length {
				t.Errorf("OTP length = %d; want %d", len(otp), tt.length)
			}

			// Every character must be a digit.
			for i, ch := range otp {
				if ch < '0' || ch > '9' {
					t.Errorf("OTP[%d] = %c; want digit [0-9]", i, ch)
				}
			}
		})
	}
}

func TestGenerateOTPUniqueness(t *testing.T) {
	// Generate many OTPs and check they're not all the same.
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		otp, err := GenerateOTP(6)
		if err != nil {
			t.Fatalf("GenerateOTP: %v", err)
		}
		seen[otp] = true
	}

	// With 6-digit OTPs (1M possibilities), 100 draws should produce
	// many distinct values. Getting fewer than 50 distinct values from
	// 100 draws would indicate a broken RNG.
	if len(seen) < 50 {
		t.Errorf("only %d distinct OTPs in 100 draws; expected high variance", len(seen))
	}
}

func TestGetDurationFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		fallback time.Duration
		want     time.Duration
	}{
		{"returns parsed duration", "5m", 10 * time.Minute, 5 * time.Minute},
		{"returns fallback when unset", "", 10 * time.Minute, 10 * time.Minute},
		{"returns fallback on invalid format", "not-a-duration", 7 * time.Second, 7 * time.Second},
		{"parses hours", "2h", time.Minute, 2 * time.Hour},
		{"parses seconds", "30s", time.Minute, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const key = "TEST_DURATION_ENV"
			os.Unsetenv(key)
			t.Cleanup(func() { os.Unsetenv(key) })

			if tt.envVal != "" {
				os.Setenv(key, tt.envVal)
			}

			got := GetDurationFromEnv(key, tt.fallback)
			if got != tt.want {
				t.Errorf("GetDurationFromEnv(%q) = %v; want %v", tt.envVal, got, tt.want)
			}
		})
	}
}
