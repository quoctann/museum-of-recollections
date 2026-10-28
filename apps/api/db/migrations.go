// Package db exposes migrations for isolated integration tests and tooling.
// Production migrations must be run explicitly using goose v3.24.1, never on API startup.
package db

import "embed"

// Migrations contains goose-compatible SQL migrations.
//
//go:embed migrations/*.sql
var Migrations embed.FS
