package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/uptrace/bun"
)

type AccountService interface {
	CreateAccount(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID, accountType model.AccountType) error
	GetAccountsByUserID(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID) ([]model.Account, error)
}
