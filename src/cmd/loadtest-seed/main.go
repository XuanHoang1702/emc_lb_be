package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// loadtest-seed tạo các user đã xác thực email + mật khẩu biết trước để k6
// có thể login -> add to cart -> checkout. Mật khẩu hash theo đúng
// utils.HashPassword(password, systemSecret) nên login chạy qua hệ thống thật.
func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, relying on system env variables")
	}

	dsn := utils.GetEnv("POSTGRES_DSN", "")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is not set")
	}
	systemSecret := utils.GetEnv("SYSTEM_SECRET", "")
	if systemSecret == "" {
		log.Fatal("SYSTEM_SECRET is not set")
	}

	count := 10
	if v := os.Getenv("LOADTEST_USERS"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &count); err != nil || count < 1 {
			log.Fatal("invalid LOADTEST_USERS (expected positive integer)")
		}
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	hash, err := utils.HashPassword("LoadTest123!", systemSecret)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	queries := sqlc.New(db)
	created := 0
	for i := 1; i <= count; i++ {
		email := fmt.Sprintf("loadtest_%03d@example.com", i)
		userName := fmt.Sprintf("loadtest_user_%03d", i)
		phone := fmt.Sprintf("091000%04d", i)

		_, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: hash,
			UserName:     userName,
			Phone:        &phone,
		})
		if err != nil {
			// Unique constraint — user đã tồn tại từ lần chạy trước, bỏ qua
			log.Printf("skip %s (may exist): %v", email, err)
			continue
		}
		if err := queries.VerifyUserEmail(ctx, email); err != nil {
			log.Printf("verify %s: %v", email, err)
		}
		created++
	}

	log.Printf("loadtest-seed: created/verified %d users (total %d)", created, count)
}
