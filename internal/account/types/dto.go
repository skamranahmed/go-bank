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
