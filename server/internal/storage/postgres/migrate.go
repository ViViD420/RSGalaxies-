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

//go:embed migrations/0002_accounts.sql
var accountsMigration string

//go:embed migrations/0003_player_bank.sql
var playerBankMigration string

func Migrate(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil { return fmt.Errorf("begin migration: %w", err) }
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(728194632)"); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	migrations := []struct { version int; sql string }{
		{version: 1, sql: initialMigration},
		{version: 2, sql: accountsMigration},
		{version: 3, sql: playerBankMigration},
	}
	for _, migration := range migrations {
		var applied bool
		if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", migration.version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration version %d: %w", migration.version, err)
		}
		if applied { continue }
		for _, statement := range strings.Split(migration.sql, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" { continue }
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply migration %d: %w", migration.version, err)
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.version); err != nil {
			return fmt.Errorf("record migration %d: %w", migration.version, err)
		}
	}
	if err := tx.Commit(); err != nil { return fmt.Errorf("commit migration: %w", err) }
	return nil
}
