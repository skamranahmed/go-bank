package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/uptrace/bun"
)

type AccountRepository interface {
	CreateAccount(requestCtx context.Context, dbExecutor bun.IDB, account *model.Account) error
	GetAccountsByUserID(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID) ([]model.Account, error)
	GetAccountByID(requestCtx context.Context, dbExecutor bun.IDB, accountID int64) (*model.Account, error)
}
