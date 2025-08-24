package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccountService interface {
	CreateAccount(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID) error
}
