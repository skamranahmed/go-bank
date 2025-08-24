package service

import (
	"context"

	"github.com/skamranahmed/go-bank/internal/user/dto"
	"github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/internal/user/repository"
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

	user := &model.User{
		Email:    email,
		Password: password, // TODO: hash this
		Username: username,
	}

	user, err := u.userRepository.CreateUser(requestCtx, dbExecutor, user)
	if err != nil {
		return nil, err
	}

	return dto.TransformToCreateUserDto(user), nil
}
