//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sqlcdn "emc_lb/src/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// runPostgres starts a real Postgres container and returns a connectable pool.
func runPostgres(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("emc_lb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	// testcontainers exposes Postgres over plain TCP on localhost; disable TLS.
	if strings.Contains(connStr, "?") {
		connStr += "&sslmode=disable"
	} else {
		connStr += "?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	// Postgres may still be accepting the boot sequence; poll until ping succeeds.
	deadline := time.Now().Add(45 * time.Second)
	for {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		pingErr := pool.Ping(pingCtx)
		cancel()
		if pingErr == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("ping postgres: %v", pingErr)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return pool
}

// runMigrations executes every *.up.sql file in the given dir (relative to the test file).
func runMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool, dir string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read migration %s: %v", f, err)
		}
		for _, stmt := range splitSQLStatements(string(raw)) {
			if _, err := pool.Exec(ctx, stmt); err != nil {
				t.Fatalf("exec migration %s: %v\nSQL:\n%s", f, err, stmt)
			}
		}
	}
}

func splitSQLStatements(sql string) []string {
	var stmts []string
	var current strings.Builder
	inDollarQuote := false

	for i := 0; i < len(sql); i++ {
		c := sql[i]

		// check for $$
		if c == '$' && i+1 < len(sql) && sql[i+1] == '$' {
			inDollarQuote = !inDollarQuote
			current.WriteByte(c)
			current.WriteByte(c)
			i++
			continue
		}

		if c == ';' && !inDollarQuote {
			s := strings.TrimSpace(current.String())
			if s != "" {
				stmts = append(stmts, s)
			}
			current.Reset()
			continue
		}

		current.WriteByte(c)
	}

	s := strings.TrimSpace(current.String())
	if s != "" {
		stmts = append(stmts, s)
	}

	return stmts
}

func TestPostgres_RBAC_SeededPermissions(t *testing.T) {
	ctx := context.Background()
	pool := runPostgres(t, ctx)
	runMigrations(t, ctx, pool, "../../internal/db/migrations")

	queries := sqlcdn.New(pool)

	rows, err := queries.GetAllRolePermissions(ctx)
	if err != nil {
		t.Fatalf("GetAllRolePermissions: %v", err)
	}

	// Seeded: 12 admin + 3 customer = 15 assignments.
	if len(rows) != 15 {
		t.Fatalf("expected 15 seeded role-permission rows, got %d", len(rows))
	}

	seen := make(map[string]bool)
	for _, rp := range rows {
		seen[rp.RoleCode+":"+rp.PermissionCode] = true
	}
	for _, key := range []string{"admin:product:read", "customer:product:read", "customer:brand:read"} {
		if !seen[key] {
			t.Fatalf("expected seeded permission %s", key)
		}
	}
}

func TestPostgres_UserRoundTrip(t *testing.T) {
	ctx := context.Background()
	pool := runPostgres(t, ctx)
	runMigrations(t, ctx, pool, "../../internal/db/migrations")

	queries := sqlcdn.New(pool)
	phone := "0912345678"
	userName := "ITest"
	now := time.Now()

	created, err := queries.CreateUser(ctx, sqlcdn.CreateUserParams{
		Email:        "itest@example.com",
		PasswordHash: "$2a$10$hash",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	
	err = queries.CreateUserProfile(ctx, sqlcdn.CreateUserProfileParams{
		UserID:   created.ID,
		UserName: &userName,
		Phone:    &phone,
	})
	if err != nil {
		t.Fatalf("CreateUserProfile: %v", err)
	}

	if created.Uuid == uuid.Nil {
		t.Fatalf("expected non-nil uuid, got %v", created.Uuid)
	}
	if created.CreatedAt.Before(now.Add(-time.Minute)) {
		t.Fatalf("created_at should be recent, got %v", created.CreatedAt)
	}

	got, err := queries.GetUserByEmail(ctx, "itest@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if got.Uuid != created.Uuid {
		t.Fatalf("expected uuid %v, got %v", created.Uuid, got.Uuid)
	}
	if got.Role != "customer" {
		t.Fatalf("default role should be customer, got %s", got.Role)
	}

	if err := queries.VerifyUserEmail(ctx, "itest@example.com"); err != nil {
		t.Fatalf("VerifyUserEmail: %v", err)
	}
	verified, err := queries.GetUserByEmail(ctx, "itest@example.com")
	if err != nil {
		t.Fatalf("re-fetch after verify: %v", err)
	}
	if !verified.EmailVerified {
		t.Fatal("expected email_verified true after VerifyUserEmail")
	}
}
