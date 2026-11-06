package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed migrations/0001_initial.sql
var initialMigration string

func Migrate(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil { return fmt.Errorf("begin migration: %w", err) }
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	var applied bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = 1)").Scan(&applied); err != nil {
		return fmt.Errorf("check migration version: %w", err)
	}
	if !applied {
		for _, statement := range strings.Split(initialMigration, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" { continue }
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply initial migration: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (1)"); err != nil {
			return fmt.Errorf("record initial migration: %w", err)
		}
	}
	if err := tx.Commit(); err != nil { return fmt.Errorf("commit migration: %w", err) }
	return nil
}
