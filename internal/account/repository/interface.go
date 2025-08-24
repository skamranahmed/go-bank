package repository

import (
	"context"

	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/uptrace/bun"
)

type AccountRepository interface {
	CreateAccount(requestCtx context.Context, dbExecutor bun.IDB, account *model.Account) error
}
