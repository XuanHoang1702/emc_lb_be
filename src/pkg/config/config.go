package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var (
	once     sync.Once
	instance *AppConfig
	initErr  error
)

// AppConfig holds all application configuration loaded from environment variables.
// It is initialized once at startup and shared read-only across the application.
type AppConfig struct {
	App      AppSettings
	Postgres PostgresSettings
	MongoDB  MongoSettings
	Redis    RedisSettings
	JWT      JWTSettings
	AWS      AWSSettings
	Mail     MailSettings
	Payment  PaymentSettings
	Search   SearchSettings
}

type AppSettings struct {
	Port         string
	Mode         string // "debug" | "release"
	SystemSecret string
	APIKey       string
	CORSOrigins  string
	LogLevel     string // "debug" | "info" | "warn" | "error"
}

type PostgresSettings struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	Timezone string
	URL      string // overrides Host/Port/User/... if set
	MaxConns int32
	MinConns int32
}

type MongoSettings struct {
	URI      string
	Database string
}

type RedisSettings struct {
	Host     string
	Port     string
	Password string
}

type SearchSettings struct {
	Host      string
	MasterKey string
}

type JWTSettings struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type AWSSettings struct {
	Region          string
	Endpoint        string // LocalStack endpoint for dev
	AccessKeyID     string
	SecretAccessKey string
	AvatarBucket    string
}

type MailSettings struct {
	Provider     string // "smtp" | "resend"
	Host         string
	Port         int
	Username     string
	Password     string
	FromEmail    string
	FromName     string
	ResendAPIKey string
	ResendFrom   string
	OTPTTl       time.Duration
}

type PaymentSettings struct {
	SepayEnv        string
	SepayMerchantID string
	SepaySecretKey  string
	SepaySuccessURL string
	SepayErrorURL   string
	SepayCancelURL  string
}

// Get returns the singleton AppConfig. Panics on first call if config is invalid.
// Call Load() explicitly in main.go for graceful error handling instead.
func Get() *AppConfig {
	once.Do(func() {
		instance, initErr = load()
	})

	if initErr != nil {
		panic(fmt.Sprintf("config: failed to load: %v", initErr))
	}

	return instance
}

// Load initializes and validates config, returning an error instead of panicking.
// Should be called once at the start of main().
func Load() (*AppConfig, error) {
	once.Do(func() {
		instance, initErr = load()
	})

	return instance, initErr
}

func load() (*AppConfig, error) {
	// Load .env file if present (no-op in production where env is set externally)
	_ = godotenv.Load()

	cfg := &AppConfig{
		App: AppSettings{
			Port:         getEnv("APP_PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			SystemSecret: getEnv("SYSTEM_SECRET", ""),
			APIKey:       getEnv("API_KEY", ""),
			CORSOrigins:  getEnv("CORS_ALLOWED_ORIGINS", "*"),
			LogLevel:     getEnv("LOG_LEVEL", "info"),
		},
		Postgres: PostgresSettings{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DB:       getEnv("POSTGRES_DB", "emc_lb"),
			Timezone: getEnv("POSTGRES_TIMEZONE", "Asia/Ho_Chi_Minh"),
			URL:      getEnv("DATABASE_URL", ""),
			// #nosec G115
			MaxConns: int32(getEnvInt("POSTGRES_MAX_CONNS", 10)),
			// #nosec G115
			MinConns: int32(getEnvInt("POSTGRES_MIN_CONNS", 2)),
		},
		MongoDB: MongoSettings{
			URI:      getEnv("MONGO_URI", ""),
			Database: getEnv("MONGO_DB", "emc_lb"),
		},
		Redis: RedisSettings{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Search: SearchSettings{
			Host:      getEnv("MEILISEARCH_HOST", "http://localhost:7700"),
			MasterKey: getEnv("MEILISEARCH_MASTER_KEY", ""),
		},
		JWT: JWTSettings{
			AccessSecret:  getEnv("ACCESS_TOKEN_SECRET", ""),
			RefreshSecret: getEnv("REFRESH_TOKEN_SECRET", ""),
			AccessTTL:     getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTTL:    getEnvDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		},
		AWS: AWSSettings{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			Endpoint:        getEnv("LOCALSTACK_ENDPOINT", ""),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", "test"),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", "test"),
			AvatarBucket:    getEnv("S3_AVATAR_BUCKET", "emc-lb-avatars"),
		},
		Mail: MailSettings{
			Provider:     getEnv("MAIL_PROVIDER", "smtp"),
			Host:         getEnv("SMTP_HOST", ""),
			Port:         getEnvInt("SMTP_PORT", 587),
			Username:     getEnv("SMTP_USERNAME", ""),
			Password:     getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("SMTP_FROM_EMAIL", ""),
			FromName:     getEnv("SMTP_FROM_NAME", "EMC LB"),
			ResendAPIKey: getEnv("RESEND_API_KEY", ""),
			ResendFrom:   getEnv("RESEND_FROM", ""),
			OTPTTl:       getEnvDuration("EMAIL_OTP_TTL", 10*time.Minute),
		},
		Payment: PaymentSettings{
			SepayEnv:        getEnv("SEPAY_ENV", "sandbox"),
			SepayMerchantID: getEnv("SEPAY_MERCHANT_ID", ""),
			SepaySecretKey:  getEnv("SEPAY_SECRET_KEY", ""),
			SepaySuccessURL: getEnv("SEPAY_SUCCESS_URL", ""),
			SepayErrorURL:   getEnv("SEPAY_ERROR_URL", ""),
			SepayCancelURL:  getEnv("SEPAY_CANCEL_URL", ""),
		},
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks critical required fields. Only fields that will cause crashes
// or security holes at runtime if missing are validated here — others are optional.
func validate(cfg *AppConfig) error {
	if cfg.App.SystemSecret == "" {
		return fmt.Errorf("SYSTEM_SECRET is required")
	}

	if cfg.MongoDB.URI == "" {
		return fmt.Errorf("MONGO_URI is required")
	}

	// JWT secrets must never fall back to known public defaults, or any
	// attacker can forge access/refresh tokens.
	if cfg.JWT.AccessSecret == "" {
		return fmt.Errorf("ACCESS_TOKEN_SECRET is required")
	}
	if cfg.JWT.RefreshSecret == "" {
		return fmt.Errorf("REFRESH_TOKEN_SECRET is required")
	}
	for name, secret := range map[string]string{
		"ACCESS_TOKEN_SECRET":  cfg.JWT.AccessSecret,
		"REFRESH_TOKEN_SECRET": cfg.JWT.RefreshSecret,
	} {
		switch secret {
		case "access-secret", "refresh-secret":
			return fmt.Errorf("%s must not use the default placeholder value", name)
		}
	}

	return nil
}

// DatabaseURL returns the full Postgres DSN, preferring DATABASE_URL env var.
func (p *PostgresSettings) DatabaseURL() string {
	if p.URL != "" {
		return p.URL
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&timezone=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.Timezone,
	)
}

// RedisAddr returns "host:port" for Redis connection.
func (r *RedisSettings) Addr() string {
	return r.Host + ":" + r.Port
}

// --- helpers ---

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}

	return d
}
