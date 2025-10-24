package repository

import (
	"context"

	"github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/skamranahmed/go-bank/internal/user/types"
	"github.com/uptrace/bun"
)

type UserRepository interface {
	CreateUser(requestCtx context.Context, dbExecutor bun.IDB, user *model.User) (*model.User, error)
	GetUser(requestCtx context.Context, dbExecutor bun.IDB, options types.UserQueryOptions) (*model.User, error)
}
