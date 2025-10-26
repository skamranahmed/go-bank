package repository

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/skamranahmed/go-bank/internal/account/types"
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

func (r *accountRepository) GetAccount(requestCtx context.Context, dbExecutor bun.IDB, options types.AccountQueryOptions) (*model.Account, error) {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	var account model.Account
	query := dbExecutor.NewSelect().Model(&account)

	// fetch only specified columns if any
	if len(options.Columns) > 0 {
		query = query.Column(options.Columns...)
	}

	// dynamically construct the query based on which fields are set
	if options.AccountID != nil {
		query = query.Where("id = ?", *options.AccountID)
	}

	// apply row locking if requested
	if options.ForUpdate {
		query = query.For("UPDATE")
	}

	err := query.Scan(requestCtx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &server.ApiError{
				HttpStatusCode: http.StatusNotFound,
				Message:        "Account not found",
			}
		}

		logger.Error(requestCtx, "Error while finding account with options: %+v, error: %+v", options, err)
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "We couldn't fetch your account at the moment. Please try again later.",
		}
	}

	return &account, nil
}
