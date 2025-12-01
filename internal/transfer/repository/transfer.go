package repository

import (
	"context"
	"net/http"

	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal/transfer/model"
	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/uptrace/bun"
)

type transferRepository struct {
	db *bun.DB
}

func NewTransferRepository(db *bun.DB) TransferRepository {
	return &transferRepository{
		db: db,
	}
}

func (r *transferRepository) CreateTransferRecord(requestCtx context.Context, dbExecutor bun.IDB, transfer *model.Transfer) (*model.Transfer, error) {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	err := dbExecutor.NewInsert().
		Model(transfer).
		Scan(requestCtx)
	if err != nil {
		logger.Error(requestCtx, "Error while performing transfer from accountID: %+v to accountID: %+v, error: %+v", transfer.FromAccountID, transfer.ToAccountID, err)
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "We couldn't transfer the money at the moment. Please try again later.",
		}
	}

	return transfer, nil
}
