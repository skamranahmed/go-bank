package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddTransferStatusEnum, downAddTransferStatusEnum)
}

func upAddTransferStatusEnum(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
		CREATE TYPE enum_transfers_status AS ENUM (
			'COMPLETED',
			'FAILED'
		);
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downAddTransferStatusEnum(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`DROP TYPE enum_transfers_status;`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
