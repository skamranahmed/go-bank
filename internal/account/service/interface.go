package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/skamranahmed/go-bank/internal/account/types"
	"github.com/uptrace/bun"
)

type AccountService interface {
	CreateAccount(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID, accountType model.AccountType) error
	GetAccountsByUserID(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID) ([]model.Account, error)
	GetAccount(requestCtx context.Context, dbExecutor bun.IDB, options types.AccountQueryOptions) (*model.Account, error)
	UpdateAccount(requestCtx context.Context, dbExecutor bun.IDB, accountID int64, options types.AccountUpdateOptions) (*model.Account, error)
	CreateTransactionRecord(requestCtx context.Context, dbExecutor bun.IDB, transaction *model.Transaction) (*model.Transaction, error)
}
