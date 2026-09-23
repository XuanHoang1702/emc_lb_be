package db_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"emc_lb/src/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func newQueries(t *testing.T) (*sqlc.Queries, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, _ := pgxmock.NewPool()
	return sqlc.New(pool), pool
}

func TestCreateUser(t *testing.T) {
	q, mock := newQueries(t)

	id := uuid.New()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	phone := "0912345678"

	mock.ExpectQuery(regexp.QuoteMeta(`
INSERT INTO users (`)).
		WithArgs(id, "john@example.com", "$2a$10$hash", "John", &phone).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "user_name", "phone", "created_at"}).
			AddRow(id, "john@example.com", "John", &phone, now))

	row, err := q.CreateUser(context.Background(), sqlc.CreateUserParams{
		ID:           id,
		Email:        "john@example.com",
		PasswordHash: "$2a$10$hash",
		UserName:     "John",
		Phone:        &phone,
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if row.ID != id {
		t.Fatalf("expected id %v, got %v", id, row.ID)
	}
	if row.CreatedAt != now {
		t.Fatalf("expected created_at %v, got %v", now, row.CreatedAt)
	}
}

func TestGetUserByEmail(t *testing.T) {
	q, mock := newQueries(t)

	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT
    id,
    email,
    password_hash,
    email_verified,
    role,
    status,
    is_banned,
    locked_until,
    failed_login_attempts
FROM users
WHERE email = $1 AND is_deleted = false LIMIT 1`)).
		WithArgs("john@example.com").
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password_hash", "email_verified", "role", "status", "is_banned", "locked_until", "failed_login_attempts"}).
			AddRow(id, "john@example.com", "$2a$10$hash", true, "customer", "active", false, time.Time{}, int32(0)))

	row, err := q.GetUserByEmail(context.Background(), "john@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail returned error: %v", err)
	}
	if row.Role != "customer" {
		t.Fatalf("expected role customer, got %s", row.Role)
	}
	if !row.EmailVerified {
		t.Fatal("expected email_verified to be true")
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	q, mock := newQueries(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT
    id,
    email,
    password_hash,
    email_verified,
    role,
    status,
    is_banned,
    locked_until,
    failed_login_attempts
FROM users
WHERE email = $1 AND is_deleted = false LIMIT 1`)).
		WithArgs("missing@example.com").
		WillReturnError(pgx.ErrNoRows)

	_, err := q.GetUserByEmail(context.Background(), "missing@example.com")
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestGetUserIDByID(t *testing.T) {
	q, mock := newQueries(t)

	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id
FROM users
WHERE id = $1 AND is_deleted = false`)).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(id))

	got, err := q.GetUserIDByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetUserIDByID returned error: %v", err)
	}
	if got != id {
		t.Fatalf("expected id %v, got %v", id, got)
	}
}

func TestSoftDeleteUserByID(t *testing.T) {
	q, mock := newQueries(t)

	id := uuid.New()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users
SET
    is_deleted = true,
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false`)).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := q.SoftDeleteUserByID(context.Background(), id); err != nil {
		t.Fatalf("SoftDeleteUserByID returned error: %v", err)
	}
}

func TestVerifyUserEmail(t *testing.T) {
	q, mock := newQueries(t)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users
SET
    email_verified = true,
    email_verified_at = NOW(),
    updated_at = NOW()
WHERE email = $1`)).
		WithArgs("john@example.com").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := q.VerifyUserEmail(context.Background(), "john@example.com"); err != nil {
		t.Fatalf("VerifyUserEmail returned error: %v", err)
	}
}

func TestUpdateUserAvatarByID(t *testing.T) {
	q, mock := newQueries(t)

	id := uuid.New()
	avatarURL := "https://s3.example.com/avatar.png"

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users
SET
    avatar_url = $2,
    updated_at = NOW()
WHERE id = $1 AND is_deleted = false`)).
		WithArgs(id, &avatarURL).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := q.UpdateUserAvatarByID(context.Background(), sqlc.UpdateUserAvatarByIDParams{
		ID:        id,
		AvatarUrl: &avatarURL,
	}); err != nil {
		t.Fatalf("UpdateUserAvatarByID returned error: %v", err)
	}
}
