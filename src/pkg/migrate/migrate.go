// Package migrate provides auto-migration support using golang-migrate.
// It applies pending SQL migrations (embedded in the binary) to the target
// database.
package migrate

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"    // embedded FS source driver

	"emc_lb/src/internal/db/migrations"
)

// Run applies all pending migrations to the database at databaseURL.
//
// Migrations are embedded at build time, so no on-disk migration directory is
// required at runtime (important for distroless container images).
//
// Behaviour:
//   - If there are no pending migrations, it returns nil (idempotent).
//   - If the migration state is "dirty" (a previous run failed mid-way), it
//     logs a warning and returns an error so the operator can fix it manually.
//   - Migration errors are wrapped with context.
//
// Call this once during application startup, before opening the query layer.
func Run(databaseURL string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migrate: init source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return fmt.Errorf("migrate: init: %w", err)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			slog.Warn("migrate: close source error", "error", srcErr)
		}

		if dbErr != nil {
			slog.Warn("migrate: close db error", "error", dbErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("migrate: no pending migrations")
			return nil
		}

		// Dirty state means a migration partially applied and failed.
		// Operator must manually fix the schema and set the version clean.
		if isDirty(err) {
			slog.Error("migrate: database is in dirty state — manual intervention required", "error", err)
			return fmt.Errorf("migrate: dirty state: %w", err)
		}

		return fmt.Errorf("migrate: up: %w", err)
	}

	version, _, _ := m.Version()
	slog.Info("migrate: applied successfully", "version", version)

	return nil
}

func isDirty(err error) bool {
	var dirtyErr migrate.ErrDirty
	return errors.As(err, &dirtyErr)
}