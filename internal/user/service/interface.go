package service

import (
	"context"

	"github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/internal/user/types"
	"github.com/uptrace/bun"
)

type UserService interface {
	CreateUser(requestCtx context.Context, dbExecutor bun.IDB, email string, password string, username string) (*types.CreateUserDto, error)
	GetUser(requestCtx context.Context, dbExecutor bun.IDB, options types.UserQueryOptions) (*model.User, error)
}
