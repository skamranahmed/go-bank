package types

import (
	"time"

	transferModel "github.com/skamranahmed/go-bank/internal/transfer/model"
)

type InternalTransferRequest struct {
	Data InternalTransferRequestData `json:"data" binding:"required"`
}

type InternalTransferRequestData struct {
	FromAccountID int64  `json:"from_account_id" binding:"required"`
	ToAccountID   int64  `json:"to_account_id" binding:"required"`
	Amount        *int64 `json:"amount" binding:"required,gt=0"`
}

type InternalTransferResponse struct {
	Data InternalTransferResponseData `json:"data"`
}

type InternalTransferResponseData struct {
	Transfer TransferDto `json:"transfer"`
}

type ExternalTransferRequest struct {
	Data ExternalTransferRequestData `json:"data" binding:"required"`
}

type ExternalTransferRequestData struct {
	Amount            *int64            `json:"amount" binding:"required,gt=0"`
	FromAccountID     int64             `json:"from_account_id" binding:"required"`
	ExternalRecipient ExternalRecipient `json:"external_recipient" binding:"required"`
}

type ExternalRecipient struct {
	AccountID int64  `json:"account_id" binding:"required,gt=0"`
	IFSCCode  string `json:"ifsc_code" binding:"required,min=1"`
	Name      string `json:"name" binding:"required,min=1"`
}

type ExternalTransferResponse struct {
	Data ExternalTransferResponseData `json:"data"`
}

type ExternalTransferResponseData struct {
	Transfer TransferDto `json:"transfer"`
}

type TransferDto struct {
	ID            string                       `json:"id"`
	CreatedAt     time.Time                    `json:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at"`
	Type          transferModel.TransferType   `json:"type"`
	FromAccountID *int64                       `json:"from_account_id"`
	ToAccountID   *int64                       `json:"to_account_id"`
	Amount        int64                        `json:"amount"`
	Status        transferModel.TransferStatus `json:"status"`
}

func TransformToTransferDto(transfer *transferModel.Transfer) *TransferDto {
	return &TransferDto{
		ID:            transfer.ID.String(),
		CreatedAt:     transfer.CreatedAt,
		UpdatedAt:     transfer.UpdatedAt,
		Type:          transfer.Type,
		FromAccountID: transfer.FromAccountID,
		ToAccountID:   transfer.ToAccountID,
		Amount:        transfer.Amount,
		Status:        transfer.Status,
	}
}
