package database

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel"
)

// RunInTransaction is a helper function to run a set of database operations within a transaction.
// It creates a new parent span for the transaction so that all operations performed within the
// transaction are grouped together and can be visualized as a single unit of work in APM traces on Kibana
func RunInTransaction(
	ctx context.Context,
	txName string,
	db *bun.DB,
	opts *sql.TxOptions,
	queryExecFunc func(ctx context.Context, tx bun.Tx) error,
) error {
	tracer := otel.Tracer("db-transaction")

	// create a new span to group the operations performed in the transaction
	txCtx, span := tracer.Start(ctx, txName)
	defer span.End()

	return db.RunInTx(txCtx, opts, queryExecFunc)
}
