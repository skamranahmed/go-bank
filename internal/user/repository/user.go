package repository

import (
	"context"
	"database/sql"
	"errors"
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &server.ApiError{
				HttpStatusCode: http.StatusNotFound,
				Message:        "User not found",
			}
		}

		logger.Error(requestCtx, "Error while finding user with options: %+v, error: %+v", options, err)
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to process request at this time. Please try again later.",
		}
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(requestCtx context.Context, dbExecutor bun.IDB, userID string, options types.UserUpdateOptions) (*model.User, error) {
	if dbExecutor == nil {
		dbExecutor = r.db
	}

	var user model.User
	query := dbExecutor.NewUpdate().Model(&user)

	// dynamically construct the update query based on which fields are set
	if options.Username != nil {
		query = query.Set("username = ?", *options.Username)
	}

	if options.HashedPassword != nil {
		query = query.Set("password = ?", *options.HashedPassword)
	}

	// always update the updated_at timestamp
	query = query.Set("updated_at = NOW()").
		Where("id = ?", userID).
		Returning("*")

	_, err := query.Exec(requestCtx)
	if err != nil {
		logger.Error(requestCtx, "Error while updating user with ID: %s, options: %+v, error: %+v", userID, options, err)
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, &server.ApiError{
				HttpStatusCode: http.StatusConflict,
				Message:        "This username is already in use. Please choose another.",
			}
		}
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to update user at this time. Please try again later.",
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
