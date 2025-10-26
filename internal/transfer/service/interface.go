package service

import (
	"context"

	"github.com/google/uuid"
	accountModel "github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/uptrace/bun"
)

type TransferService interface {
	CreateInternalTransfer(requestCtx context.Context, dbExecutor bun.IDB, senderUserID uuid.UUID, fromAccountID, toAccountID, transferAmount int64) (*accountModel.Transaction, error)
}
