//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// TestMigrationLifecycle runs a full Up and Down migration to ensure schema integrity.
func TestMigrationLifecycle(t *testing.T) {
	ctx := context.Background()

	// 1. Fresh DB Initialization
	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("emc_lb_migration_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}
	if strings.Contains(connStr, "?") {
		connStr += "&sslmode=disable"
	} else {
		connStr += "?sslmode=disable"
	}

	// 2. Locate migrations folder
	migrationsPath := "file://../../internal/db/migrations"

	// 3. Migrate UP
	m, err := migrate.New(migrationsPath, connStr)
	if err != nil {
		t.Fatalf("failed to initialize migrate: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up failed: %v", err)
	}

	// Basic validation (if Up didn't panic and returned no error, schema is applied)
	version, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("failed to get migration version: %v", err)
	}
	if dirty {
		t.Fatalf("database is in a dirty state after migrate up")
	}
	t.Logf("Migrated UP successfully to version %d", version)

	// 4. Rollback (Down) by 1 step
	err = m.Steps(-1)
	if err != nil {
		t.Fatalf("migrate down 1 step failed: %v", err)
	}

	versionAfterDown, _, _ := m.Version()
	t.Logf("Migrated DOWN successfully to version %d", versionAfterDown)

	// 5. Migrate UP again to ensure the down script was clean
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up (second time) failed: %v", err)
	}
}
