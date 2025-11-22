package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateTransfersTable, downCreateTransfersTable)
}

func upCreateTransfersTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
		CREATE TABLE transfers (
			id UUID PRIMARY KEY NOT NULL DEFAULT GEN_RANDOM_UUID(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			type enum_transfers_type NOT NULL,
			from_account_id BIGINT REFERENCES accounts(id),
			to_account_id BIGINT REFERENCES accounts(id),
			amount BIGINT NOT NULL CHECK (amount > 0),
			status enum_transfers_status NOT NULL
		);

		COMMENT ON COLUMN transfers.amount IS 'Amount involved in the transfer, in the lowest currency unit i.e paise for INR';
		COMMENT ON COLUMN transfers.status IS 'Status of the transfer, can be COMPLETED or FAILED';
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downCreateTransfersTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`DROP TABLE transfers;`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
