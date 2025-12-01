package service

import (
	"context"

	"github.com/google/uuid"
	transferModel "github.com/skamranahmed/go-bank/internal/transfer/model"
	"github.com/uptrace/bun"
)

type TransferService interface {
	CreateInternalTransfer(requestCtx context.Context, dbExecutor bun.IDB, senderUserID uuid.UUID, fromAccountID, toAccountID, transferAmount int64) (*transferModel.Transfer, error)
}
