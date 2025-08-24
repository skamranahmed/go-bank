package repository

import (
	"context"
	"fmt"
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
		logger.Warnf("error while creating new account for userID: %+v, error: %+v", account.UserID, err)
		return &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        fmt.Sprintf("something went wrong while creating the account for user id: %v", account.UserID),
		}
	}

	return nil
}
