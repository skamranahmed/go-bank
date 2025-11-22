package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddTransferTypeEnum, downAddTransferTypeEnum)
}

func upAddTransferTypeEnum(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
		CREATE TYPE enum_transfers_type AS ENUM (
			'INTERNAL', -- transfer between accounts within our own bank (go-bank)
			'EXTERNAL_INBOUND', -- transfer coming from an external bank to an account in our bank (go-bank)
			'EXTERNAL_OUTBOUND' -- transfer going from an account in our bank (go-bank) to an external bank
		);
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downAddTransferTypeEnum(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`DROP TYPE enum_transfers_type`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
