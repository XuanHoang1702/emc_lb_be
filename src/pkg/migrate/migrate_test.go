package migrate

import (
	"errors"
	"io/fs"
	"testing"

	"emc_lb/src/internal/db/migrations"

	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
)

func TestEmbeddedMigrationsSource(t *testing.T) {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("iofs.New: %v", err)
	}
	defer func() { _ = source.Close() }()

	// Walk the migration versions via the source.Driver interface. The
	// embed.go file must be filtered out by golang-migrate's filename pattern.
	var versions []uint
	version, err := source.First()
	for ; err == nil; version, err = source.Next(version) {
		versions = append(versions, version)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist at end of migrations, got %v", err)
	}

	if len(versions) < 1 {
		t.Fatalf("expected at least 1 embedded migration, got %d (%v)", len(versions), versions)
	}
	for i, v := range versions {
		if v != uint(i+1) {
			t.Fatalf("expected sequential version %d at index %d, got %d", i+1, i, v)
		}
	}
}
