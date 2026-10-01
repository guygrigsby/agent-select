// Package collector owns the Postgres authority for the educated routing
// delivery: numbered migrations instantiated from the contract specs and the
// mutation ports defined in docs/specs/delivery-storage.md.
package collector

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// MigrationFiles returns the numbered migration names in application order.
func MigrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for i, n := range names {
		if !strings.HasPrefix(n, fmt.Sprintf("%04d", i+1)) {
			return nil, fmt.Errorf("migration %s is not sequential at position %d", n, i+1)
		}
	}
	return names, nil
}

// Migrate applies every numbered migration that is not yet recorded. It is
// idempotent: applied migrations are skipped by name and content hash, and a
// recorded hash that no longer matches the embedded file is an error, never a
// silent re-application of different bytes.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			content_hash BYTEA NOT NULL CHECK (octet_length(content_hash) = 32),
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	names, err := MigrationFiles()
	if err != nil {
		return err
	}
	for _, name := range names {
		body, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		var recorded []byte
		err = pool.QueryRow(ctx,
			`SELECT content_hash FROM schema_migrations WHERE name = $1`, name).
			Scan(&recorded)
		switch {
		case err == nil:
			if hex.EncodeToString(recorded) != hex.EncodeToString(sum[:]) {
				return fmt.Errorf("migration %s content changed after being applied", name)
			}
			continue
		case err.Error() == "no rows in result set":
			// not yet applied
		default:
			return err
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (name, content_hash) VALUES ($1, $2)`,
			name, sum[:]); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

// AppliedMigrations returns the recorded migration names in order.
func AppliedMigrations(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, `SELECT name FROM schema_migrations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}
