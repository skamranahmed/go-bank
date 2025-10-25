package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
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

func (r *accountRepository) GetAccountsByUserID(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID) ([]model.Account, error) {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	var accounts []model.Account
	err := dbExecutor.NewSelect().
		Model(&accounts).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Scan(requestCtx)
	if err != nil {
		logger.Error(requestCtx, "Error while fetching accounts for userID: %+v, error: %+v", userID, err)
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "We couldn't fetch your accounts at the moment. Please try again later.",
		}
	}

	return accounts, nil
}

func (r *accountRepository) GetAccountByID(requestCtx context.Context, dbExecutor bun.IDB, accountID int64) (*model.Account, error) {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	var account model.Account
	err := dbExecutor.NewSelect().
		Model(&account).
		Where("id = ?", accountID).
		Scan(requestCtx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &server.ApiError{
				HttpStatusCode: http.StatusNotFound,
				Message:        "Account not found",
			}
		}

		logger.Error(requestCtx, "Error while fetching account for accountID: %+v, error: %+v", accountID, err)
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "We couldn't fetch your account at the moment. Please try again later.",
		}
	}

	return &account, nil
}
