package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddExternalPartyColumnToTransfersTable, downAddExternalPartyColumnToTransfersTable)
}

func upAddExternalPartyColumnToTransfersTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
		ALTER TABLE transfers
		ADD COLUMN external_party JSONB DEFAULT NULL;
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	_, err = tx.Exec(`
		ALTER TABLE transfers
		ADD CONSTRAINT check_external_party
		CHECK (
			(type IN ('EXTERNAL_OUTBOUND', 'EXTERNAL_INBOUND') AND external_party IS NOT NULL) OR
			(type NOT IN ('EXTERNAL_OUTBOUND', 'EXTERNAL_INBOUND') AND external_party IS NULL)
		);
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downAddExternalPartyColumnToTransfersTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`
		ALTER TABLE transfers
		DROP COLUMN external_party;
	`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
