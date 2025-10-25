package service

import (
	"context"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/internal/user/repository"
	"github.com/skamranahmed/go-bank/internal/user/types"
	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/uptrace/bun"
)

type userService struct {
	db             *bun.DB
	userRepository repository.UserRepository
}

func NewUserService(db *bun.DB, userRepository repository.UserRepository) UserService {
	return &userService{
		db:             db,
		userRepository: userRepository,
	}
}

func (u *userService) CreateUser(requestCtx context.Context, dbExecutor bun.IDB, email string, password string, username string) (*types.CreateUserDto, error) {
	if dbExecutor == nil {
		dbExecutor = u.db
	}

	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		logger.Error(requestCtx, "Error hashing the password, error: %v", err)
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to process your request. Please try again later.",
		}
	}

	user := &model.User{
		Email:    email,
		Password: hashedPassword,
		Username: username,
	}

	user, err = u.userRepository.CreateUser(requestCtx, dbExecutor, user)
	if err != nil {
		return nil, err
	}

	return types.TransformToCreateUserDto(user), nil
}

func (u *userService) GetUser(requestCtx context.Context, dbExecutor bun.IDB, options types.UserQueryOptions) (*model.User, error) {
	return u.userRepository.GetUser(requestCtx, dbExecutor, options)
}

func (s *userService) UpdateUser(requestCtx context.Context, dbExecutor bun.IDB, userID string, options types.UserUpdateOptions) (*model.User, error) {
	if dbExecutor == nil {
		dbExecutor = s.db
	}

	// verify user exists
	_, err := s.userRepository.GetUser(requestCtx, dbExecutor, types.UserQueryOptions{
		ID: &userID,
	})
	if err != nil {
		return nil, err
	}

	// update user
	updatedUser, err := s.userRepository.UpdateUser(requestCtx, dbExecutor, userID, options)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *userService) UpdatePassword(requestCtx context.Context, dbExecutor bun.IDB, userID string, currentPassword string, newPassword string) error {
	if dbExecutor == nil {
		dbExecutor = s.db
	}

	// get user with password field
	user, err := s.userRepository.GetUser(requestCtx, dbExecutor, types.UserQueryOptions{
		ID:      &userID,
		Columns: []string{"id", "password"},
	})
	if err != nil {
		return err
	}

	// verify current password
	doesPasswordMatch, err := argon2id.ComparePasswordAndHash(currentPassword, user.Password)
	if err != nil {
		logger.Error(requestCtx, "Error comparing password and hash, error: %v", err)
		return &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to process your request. Please try again later.",
		}
	}

	if !doesPasswordMatch {
		return &server.ApiError{
			HttpStatusCode: http.StatusUnauthorized,
			Message:        "Current password is incorrect",
		}
	}

	// hash new password
	hashedPassword, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
	if err != nil {
		logger.Error(requestCtx, "Error hashing the new password, error: %v", err)
		return &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to process your request. Please try again later.",
		}
	}

	updateOptions := types.UserUpdateOptions{
		HashedPassword: &hashedPassword,
	}
	_, err = s.userRepository.UpdateUser(requestCtx, dbExecutor, userID, updateOptions)
	if err != nil {
		return err
	}

	return nil
}
