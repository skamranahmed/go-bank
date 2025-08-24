package service

import (
	"context"

	"github.com/skamranahmed/go-bank/internal/user/dto"
	"github.com/uptrace/bun"
)

type UserService interface {
	CreateUser(requestCtx context.Context, dbExecutor bun.IDB, email string, password string, username string) (*dto.CreateUserDto, error)
}
