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
	UpdateUser(requestCtx context.Context, dbExecutor bun.IDB, userID string, options types.UserUpdateOptions) (*model.User, error)
	UpdatePassword(requestCtx context.Context, dbExecutor bun.IDB, userID string, currentPassword string, newPassword string) error
}
