package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Transaction struct {
	bun.BaseModel `bun:"table:transactions"`

	ID        uuid.UUID `bun:"id,pk,notnull,type:uuid,default:gen_random_uuid()"`
	CreatedAt time.Time `bun:"created_at,notnull,default:now()"`

	// foreign key to "accounts" table
	AccountID int64    `bun:"account_id,notnull"`
	Account   *Account `bun:"rel:belongs-to,join:account_id=id"`

	// Amount is stored in the smallest currency unit (paise for INR)
	Amount int64 `bun:"amount,notnull"`

	// BalanceAfter is the account balance after this transaction, stored in the smallest currency unit (paise for INR)
	BalanceAfter int64 `bun:"balance_after,notnull"`

	// Type of transaction: DEBIT, CREDIT
	Type TransactionType `bun:"type,notnull"`
}

type TransactionType string

const (
	Debit  TransactionType = "DEBIT"
	Credit TransactionType = "CREDIT"
)
