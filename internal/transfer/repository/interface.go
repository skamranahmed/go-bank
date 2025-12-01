package repository

import (
	"context"

	"github.com/skamranahmed/go-bank/internal/transfer/model"
	"github.com/uptrace/bun"
)

type TransferRepository interface {
	CreateTransferRecord(requestCtx context.Context, dbExecutor bun.IDB, transfer *model.Transfer) (*model.Transfer, error)
}
