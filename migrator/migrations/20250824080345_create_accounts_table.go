package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateAccountsTable, downCreateAccountsTable)
}

func upCreateAccountsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
		CREATE TABLE accounts (
			id BIGINT PRIMARY KEY CHECK (id BETWEEN 1000000000 AND 999999999999999),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			user_id UUID NOT NULL REFERENCES users(id),
			balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
		);

		COMMENT ON COLUMN "accounts"."id" IS 'Used as customer-facing account identifier, can also be called account number. It must be 10-15 digits.';
		COMMENT ON COLUMN "accounts"."balance" IS 'Balance stored in the smallest currency unit (paise for INR)';
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downCreateAccountsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`DROP TABLE accounts`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
