package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddUniqueConstraintOnUseridAndTypeOnAccountsTable, downAddUniqueConstraintOnUseridAndTypeOnAccountsTable)
}

func upAddUniqueConstraintOnUseridAndTypeOnAccountsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	logMigrationStatus("⬆️ Applying migration")

	_, err := tx.Exec(`
		ALTER TABLE accounts 
		ADD CONSTRAINT accounts_user_id_type_unique UNIQUE (user_id, type);
	`)
	if err != nil {
		logMigrationStatus("❌ Applying migration failed")
		return err
	}

	logMigrationStatus("✅ Migration applied")
	return nil
}

func downAddUniqueConstraintOnUseridAndTypeOnAccountsTable(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	logMigrationStatus("⬇️ Rolling back migration")

	_, err := tx.Exec(`ALTER TABLE accounts DROP CONSTRAINT accounts_user_id_type_unique`)
	if err != nil {
		logMigrationStatus("❌ Rollback failed")
		return err
	}

	logMigrationStatus("✅ Rollback done")
	return nil
}
