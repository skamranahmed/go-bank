package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/internal/user/model"
	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel `bun:"table:accounts"`

	// ID is used as customer-facing account identifier, can also be called account number. It must be 10-15 digits.
	ID int64 `bun:"id,pk,notnull"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// foreign key to "users" table
	UserID uuid.UUID   `bun:"user_id,notnull,type:uuid,unique:accounts_user_id_type_unique"`
	User   *model.User `bun:"rel:belongs-to,join:user_id=id"`

	// Balance is stored in the smallest currency unit (paise for INR)
	Balance int64 `bun:"balance,notnull,default:0"`

	// Type of bank account: SAVINGS_ACCOUNT, CURRENT_ACCOUNT
	Type AccountType `bun:"type,notnull,unique:accounts_user_id_type_unique,default:'SAVINGS_ACCOUNT'"`
}

type AccountType string

const (
	SavingsAccount AccountType = "SAVINGS_ACCOUNT"
	CurrentAccount AccountType = "CURRENT_ACCOUNT"
)
