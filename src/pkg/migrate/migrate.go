// Package migrate provides auto-migration support using golang-migrate.
// It applies pending SQL migrations from a directory to the target database.
package migrate

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // file source driver
)

// Run applies all pending migrations from migrationsDir to the database
// at databaseURL.
//
// Behaviour:
//   - If there are no pending migrations, it returns nil (idempotent).
//   - If the migration state is "dirty" (a previous run failed mid-way), it
//     logs a warning and returns an error so the operator can fix it manually.
//   - Migration errors are wrapped with context.
//
// Call this once during application startup, before opening the query layer.
func Run(databaseURL, migrationsDir string) error {
	sourceURL := fmt.Sprintf("file://%s", migrationsDir)

	m, err := migrate.New(sourceURL, databaseURL)
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

// Version returns the current applied migration version and whether the
// database is in a dirty (incomplete) migration state.
func Version(databaseURL, migrationsDir string) (uint, bool, error) {
	sourceURL := fmt.Sprintf("file://%s", migrationsDir)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return 0, false, fmt.Errorf("migrate: init: %w", err)
	}

	defer func() {
		_, _ = m.Close()
	}()

	return m.Version()
}

func isDirty(err error) bool {
	var dirtyErr migrate.ErrDirty
	return errors.As(err, &dirtyErr)
}
