package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateTransactionsTable, downCreateTransactionsTable)
}

func upCreateTransactionsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
		CREATE TABLE transactions (
			id UUID PRIMARY KEY NOT NULL DEFAULT GEN_RANDOM_UUID(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			account_id BIGINT NOT NULL REFERENCES accounts(id),
			amount BIGINT NOT NULL CHECK (amount > 0),
			balance_after BIGINT NOT NULL CHECK (balance_after >= 0),
			type enum_transactions_type NOT NULL
		);

		COMMENT ON COLUMN transactions.amount IS 'Amount involved in the transaction, in the lowest currency unit i.e paise for INR';
		COMMENT ON COLUMN transactions.balance_after IS 'Account balance after the transaction, in the lowest currency unit i.e paise for INR';
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downCreateTransactionsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`DROP TABLE transactions`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
