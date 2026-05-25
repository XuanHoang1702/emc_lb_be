package utils

import "fmt"

func BuildDatabaseURL() string {
	if value := GetEnv("DATABASE_URL", ""); value != "" {
		return value
	}

	host := GetEnv("POSTGRES_HOST", "localhost")
	port := GetEnv("POSTGRES_PORT", "5432")
	user := GetEnv("POSTGRES_USER", "postgres")
	password := GetEnv("POSTGRES_PASSWORD", "postgres")
	dbName := GetEnv("POSTGRES_DB", "emc_lb")
	timezone := GetEnv("POSTGRES_TIMEZONE", "Asia/Ho_Chi_Minh")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&timezone=%s", user, password, host, port, dbName, timezone)
}
