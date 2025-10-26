package types

import (
	accountTypes "github.com/skamranahmed/go-bank/internal/account/types"
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
	Transaction accountTypes.TransactionDto `json:"transaction"`
}
