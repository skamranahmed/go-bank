package service

import (
	"context"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal/user/dto"
	"github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/internal/user/repository"
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

func (u *userService) CreateUser(requestCtx context.Context, dbExecutor bun.IDB, email string, password string, username string) (*dto.CreateUserDto, error) {
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

	return dto.TransformToCreateUserDto(user), nil
}
