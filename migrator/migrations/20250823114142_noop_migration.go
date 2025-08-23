package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

// This file defines a "noop" (no-operation) migration.
// It serves as a placeholder or test migration. Running it will succeed
// but it will not make any changes to the database schema.
// Useful for validating that the migration framework is wired correctly
// or to intentionally create a versioned checkpoint in migration history.

func init() {
	// Register the migration with Goose.
	// Both up and down simply return nil (do nothing).
	goose.AddMigrationContext(upNoopMigration, downNoopMigration)
}

func upNoopMigration(ctx context.Context, tx *sql.Tx) error {
	// This code runs when applying the migration (goose up).
	// Intentionally does nothing.
	return nil
}

func downNoopMigration(ctx context.Context, tx *sql.Tx) error {
	// This code runs when rolling back the migration (goose down).
	// Intentionally does nothing.
	return nil
}
