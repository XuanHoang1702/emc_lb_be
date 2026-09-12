package utils

import (
	"os"
	"testing"
)

func TestBuildMongoURI(t *testing.T) {
	tests := []struct {
		name       string
		envVars    map[string]string
		want       string
		wantErr    bool
		errContain string
	}{
		{
			name:    "returns MONGO_URI when set",
			envVars: map[string]string{"MONGO_URI": "mongodb://localhost:27017"},
			want:    "mongodb://localhost:27017",
		},
		{
			name:    "falls back to MONGO_INITDB_ROOT_URL",
			envVars: map[string]string{"MONGO_INITDB_ROOT_URL": "mongodb://root:pass@host:27017"},
			want:    "mongodb://root:pass@host:27017",
		},
		{
			name:    "MONGO_URI takes priority over MONGO_INITDB_ROOT_URL",
			envVars: map[string]string{"MONGO_URI": "mongodb://primary", "MONGO_INITDB_ROOT_URL": "mongodb://fallback"},
			want:    "mongodb://primary",
		},
		{
			name:       "errors when neither is set",
			envVars:    map[string]string{},
			wantErr:    true,
			errContain: "MONGO_URI or MONGO_INITDB_ROOT_URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars.
			os.Unsetenv("MONGO_URI")
			os.Unsetenv("MONGO_INITDB_ROOT_URL")
			t.Cleanup(func() {
				os.Unsetenv("MONGO_URI")
				os.Unsetenv("MONGO_INITDB_ROOT_URL")
			})

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			got, err := BuildMongoURI()

			if tt.wantErr {
				if err == nil {
					t.Fatal("BuildMongoURI() = nil error; want error")
				}
				if tt.errContain != "" && !containsSubstring(err.Error(), tt.errContain) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("BuildMongoURI() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("BuildMongoURI() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestGetMongoDatabaseName(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   string
	}{
		{"returns env value when set", "custom_db", "custom_db"},
		{"returns default when unset", "", "emc_lb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("MONGO_DB")
			t.Cleanup(func() { os.Unsetenv("MONGO_DB") })

			if tt.envVal != "" {
				os.Setenv("MONGO_DB", tt.envVal)
			}

			got := GetMongoDatabaseName()
			if got != tt.want {
				t.Errorf("GetMongoDatabaseName() = %q; want %q", got, tt.want)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
