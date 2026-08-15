package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlcdn "emc_lb/src/internal/db/sqlc"
	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

var (
	userFindSQL = regexp.QuoteMeta(`SELECT
    id,
    email,
    password_hash,
    email_verified,
    role
FROM users
WHERE email = $1 AND is_deleted = false`)
	userCreateSQL = regexp.QuoteMeta(`
INSERT INTO users (`)
	userVerifySQL = regexp.QuoteMeta(`UPDATE users
SET
    email_verified = true,
    email_verified_at = NOW(),
    updated_at = NOW()
WHERE email = $1`)
	userDeleteSQL = regexp.QuoteMeta(`UPDATE users
SET
    is_deleted = true,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE email = $1 AND is_deleted = false`)
	userAvatarSQL = regexp.QuoteMeta(`UPDATE users
SET
    avatar_url = $2,
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false`)
)

func TestUserRepo_GetByEmail(t *testing.T) {
	pool, _ := pgxmock.NewPool()
	repo := repository.NewUserRepository(sqlcdn.New(pool))

	id := uuid.New()
	pool.ExpectQuery(userFindSQL).
		WithArgs("john@example.com").
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password_hash", "email_verified", "role"}).
			AddRow(id, "john@example.com", "$2a$10$hash", true, "customer"))

	user, err := repo.GetByEmail(context.Background(), "john@example.com")
	if err != nil {
		t.Fatalf("GetByEmail returned error: %v", err)
	}
	if user.Email != "john@example.com" {
		t.Fatalf("expected email john@example.com, got %s", user.Email)
	}
	if user.Role != "customer" {
		t.Fatalf("expected role customer, got %s", user.Role)
	}
	if !user.EmailVerified {
		t.Fatal("expected email_verified to be true")
	}
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	pool, _ := pgxmock.NewPool()
	repo := repository.NewUserRepository(sqlcdn.New(pool))

	pool.ExpectQuery(userFindSQL).
		WithArgs("missing@example.com").
		WillReturnError(pgx.ErrNoRows)

	_, err := repo.GetByEmail(context.Background(), "missing@example.com")
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestUserRepo_Create(t *testing.T) {
	pool, _ := pgxmock.NewPool()
	repo := repository.NewUserRepository(sqlcdn.New(pool))

	id := uuid.New()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	phone := "0912345678"

	pool.ExpectQuery(userCreateSQL).
		WithArgs(id, "john@example.com", "$2a$10$hash", "John", &phone).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "user_name", "phone", "created_at"}).
			AddRow(id, "john@example.com", "John", &phone, now))

	user, err := repo.Create(context.Background(), entities.User{
		ID:           id,
		Email:        "john@example.com",
		UserName:     "John",
		Phone:        &phone,
		PasswordHash: "$2a$10$hash",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.ID != id {
		t.Fatalf("expected id %v, got %v", id, user.ID)
	}
	if !user.CreatedAt.Equal(now) {
		t.Fatalf("expected created_at %v, got %v", now, user.CreatedAt)
	}
}

func TestUserRepo_VerifyEmail(t *testing.T) {
	pool, _ := pgxmock.NewPool()
	repo := repository.NewUserRepository(sqlcdn.New(pool))

	pool.ExpectExec(userVerifySQL).
		WithArgs("john@example.com").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := repo.VerifyEmail(context.Background(), "john@example.com"); err != nil {
		t.Fatalf("VerifyEmail returned error: %v", err)
	}
}

func TestUserRepo_SoftDeleteByEmail(t *testing.T) {
	pool, _ := pgxmock.NewPool()
	repo := repository.NewUserRepository(sqlcdn.New(pool))

	pool.ExpectExec(userDeleteSQL).
		WithArgs("john@example.com").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := repo.SoftDeleteByEmail(context.Background(), "john@example.com"); err != nil {
		t.Fatalf("SoftDeleteByEmail returned error: %v", err)
	}
}

func TestUserRepo_UpdateAvatarByID(t *testing.T) {
	pool, _ := pgxmock.NewPool()
	repo := repository.NewUserRepository(sqlcdn.New(pool))

	id := uuid.New()
	avatarURL := "https://s3.example.com/u.png"

	pool.ExpectExec(userAvatarSQL).
		WithArgs(id, &avatarURL).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := repo.UpdateAvatarByID(context.Background(), id, avatarURL); err != nil {
		t.Fatalf("UpdateAvatarByID returned error: %v", err)
	}
}
