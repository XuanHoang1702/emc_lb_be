// Package migrations embeds the SQL migration files into the binary so the
// application can run migrations without requiring the source tree to be
// present (e.g. inside a distroless container image).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
