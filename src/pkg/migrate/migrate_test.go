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

	if len(versions) != 3 {
		t.Fatalf("expected 3 embedded migrations, got %d (%v)", len(versions), versions)
	}
	if versions[0] != 1 || versions[1] != 2 || versions[2] != 3 {
		t.Fatalf("expected versions [1 2 3], got %v", versions)
	}
}
