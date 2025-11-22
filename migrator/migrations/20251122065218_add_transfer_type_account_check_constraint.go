package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddTransferTypeAccountCheckConstraint, downAddTransferTypeAccountCheckConstraint)
}

func upAddTransferTypeAccountCheckConstraint(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
    	ALTER TABLE transfers
		ADD CONSTRAINT check_transfer_type_accounts
		CHECK (
			(type = 'INTERNAL' AND from_account_id IS NOT NULL AND to_account_id IS NOT NULL) OR
			(type = 'EXTERNAL_INBOUND' AND from_account_id IS NULL AND to_account_id IS NOT NULL) OR
			(type = 'EXTERNAL_OUTBOUND' AND from_account_id IS NOT NULL AND to_account_id IS NULL)
		);
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downAddTransferTypeAccountCheckConstraint(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`ALTER TABLE transfers DROP CONSTRAINT check_transfer_type_accounts;`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
