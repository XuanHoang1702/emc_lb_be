package utils

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envVal   string // if non-empty, set this env before test
		want     string
	}{
		{"returns env value when set", "TEST_GET_ENV_1", "fallback", "actual", "actual"},
		{"returns fallback when unset", "TEST_GET_ENV_UNSET_KEY", "default_val", "", "default_val"},
		{"returns fallback for empty string env", "TEST_GET_ENV_EMPTY", "fb", "", "fb"},
		{"returns empty fallback when both empty", "TEST_GET_ENV_BOTH_EMPTY", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up after test.
			t.Cleanup(func() { os.Unsetenv(tt.key) })

			if tt.envVal != "" {
				os.Setenv(tt.key, tt.envVal)
			}

			got := GetEnv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("GetEnv(%q, %q) = %q; want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}
