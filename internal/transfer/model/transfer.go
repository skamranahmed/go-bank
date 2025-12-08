package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/uptrace/bun"
)

type Transfer struct {
	bun.BaseModel `bun:"table:transfers"`

	ID        uuid.UUID `bun:"id,pk,notnull,type:uuid,default:gen_random_uuid()"`
	CreatedAt time.Time `bun:"created_at,notnull,default:now()"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:now()"`

	Type TransferType `bun:"type,notnull"`

	FromAccountID *int64         `bun:"from_account_id"` // nullable for EXTERNAL_INBOUND transfers
	FromAccount   *model.Account `bun:"rel:belongs-to,join:from_account_id=id"`
	ToAccountID   *int64         `bun:"to_account_id"` // nullable for EXTERNAL_OUTBOUND transfers
	ToAccount     *model.Account `bun:"rel:belongs-to,join:to_account_id=id"`

	// Amount is stored in the smallest currency unit (paise for INR)
	Amount int64 `bun:"amount,notnull"`

	Status TransferStatus `bun:"status,notnull"`

	// Fields:
	// 	"ifsc_code": "XXXX" (the ifsc code of the external bank)
	// 	"account_id": "XXXX" (the account number of the sender or recipient of the external bank)
	// 	"name": "Sender name or recipient name"
	ExternalParty map[string]interface{} `bun:"external_party,default:null"` // nullable, holds info about external party for EXTERNAL_INBOUND and EXTERNAL_OUTBOUND transfers
}

type TransferType string

const (
	// Transfer between accounts within our own bank (go-bank)
	TransferTypeInternal TransferType = "INTERNAL"
	// Transfer coming from an external bank to an account in our bank (go-bank)
	TransferTypeExternalInbound TransferType = "EXTERNAL_INBOUND"
	// Transfer going from an account in our bank (go-bank) to an external bank
	TransferTypeExternalOutbound TransferType = "EXTERNAL_OUTBOUND"
)

type TransferStatus string

const (
	TransferStatusCompleted TransferStatus = "COMPLETED"
	TransferStatusFailed    TransferStatus = "FAILED"
)
