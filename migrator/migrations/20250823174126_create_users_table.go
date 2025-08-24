package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateUsersTable, downCreateUsersTable)
}

func upCreateUsersTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		CREATE TABLE "users" (
			id UUID PRIMARY KEY NOT NULL DEFAULT GEN_RANDOM_UUID(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			username VARCHAR(20) NOT NULL UNIQUE CHECK (username != ''),
			password VARCHAR(255) NOT NULL CHECK (password != ''),
			email VARCHAR(100) NOT NULL UNIQUE CHECK (email != '')
		)
	`)
	return err
}

func downCreateUsersTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec(`DROP TABLE "users"`)
	return err
}
