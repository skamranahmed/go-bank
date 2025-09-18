package repository

import (
	"context"
	"net/http"

	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/uptrace/bun"
)

type accountRepository struct {
	db *bun.DB
}

func NewAccountRepository(db *bun.DB) AccountRepository {
	return &accountRepository{
		db: db,
	}
}

func (r *accountRepository) CreateAccount(requestCtx context.Context, dbExecutor bun.IDB, account *model.Account) error {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	err := dbExecutor.NewInsert().
		Model(account).
		Scan(requestCtx)
	if err != nil {
		logger.Error(requestCtx, "Error while creating new account for userID: %+v, error: %+v", account.UserID, err)
		return &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "We couldn't create your account at the moment. Please try again later.",
		}
	}

	return nil
}
