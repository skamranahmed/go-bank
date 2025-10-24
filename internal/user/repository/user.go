package repository

import (
	"context"
	"net/http"
	"strings"

	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/internal/user/types"
	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/uptrace/bun"
)

type userRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateUser(requestCtx context.Context, dbExecutor bun.IDB, user *model.User) (*model.User, error) {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	err := dbExecutor.NewInsert().
		Model(user).
		Returning("*").
		Scan(requestCtx)
	if err != nil {
		logger.Error(requestCtx, "Error while creating new user record, error: %+v", err)
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, &server.ApiError{
				HttpStatusCode: http.StatusConflict,
				Message:        "This username or email is already in use. Please choose another.",
			}
		}
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to create user at this time. Please try again later.",
		}
	}

	safeUser := &model.User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		// Password omitted
	}

	return safeUser, nil
}

func (r *userRepository) GetUser(requestCtx context.Context, dbExecutor bun.IDB, options types.UserQueryOptions) (*model.User, error) {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	var user model.User
	query := dbExecutor.NewSelect().Model(&user)

	// fetch only specified columns if any
	if len(options.Columns) > 0 {
		query = query.Column(options.Columns...)
	}

	// dynamically construct the query based on which fields are set
	if options.Username != nil {
		query = query.Where("username = ?", *options.Username)
	}
	if options.Email != nil {
		query = query.Where("email = ?", *options.Email)
	}
	if options.ID != nil {
		query = query.Where("id = ?", *options.ID)
	}

	err := query.Scan(requestCtx)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, err
		}

		logger.Error(requestCtx, "Error while finding user with options: %+v, error: %+v", options, err)
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to process request at this time. Please try again later.",
		}
	}

	return &user, nil
}
