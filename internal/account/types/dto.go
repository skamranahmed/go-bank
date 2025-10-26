package types

import (
	"time"

	"github.com/skamranahmed/go-bank/internal/account/model"
)

type AccountDto struct {
	ID        int64             `json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	UserID    string            `json:"user_id"`
	Balance   int64             `json:"balance"`
	Type      model.AccountType `json:"type"`
}

type GetAccountsResponse struct {
	Data []AccountDto `json:"data"`
}

type GetAccountByIDResponse struct {
	Data AccountDto `json:"data"`
}

func TransformToAccountDto(account *model.Account) *AccountDto {
	return &AccountDto{
		ID:        account.ID,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
		UserID:    account.UserID.String(),
		Balance:   account.Balance,
		Type:      account.Type,
	}
}

func TransformToAccountDtoList(accounts []model.Account) []AccountDto {
	accountDtos := make([]AccountDto, 0, len(accounts))
	for _, account := range accounts {
		accountDtos = append(accountDtos, *TransformToAccountDto(&account))
	}
	return accountDtos
}

type TransactionDto struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	AccountID    int64     `json:"account_id"`
	Amount       int64     `json:"amount"`
	Type         string    `json:"type"`
	BalanceAfter int64     `json:"balance_after"`
}

func TransformToTransactionDto(transaction *model.Transaction) *TransactionDto {
	return &TransactionDto{
		ID:           transaction.ID.String(),
		CreatedAt:    transaction.CreatedAt,
		AccountID:    transaction.AccountID,
		Amount:       transaction.Amount,
		Type:         string(transaction.Type),
		BalanceAfter: transaction.BalanceAfter,
	}
}
