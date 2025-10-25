package service

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/internal/account/model"
	"github.com/skamranahmed/go-bank/internal/account/repository"
	"github.com/uptrace/bun"
)

type accountService struct {
	db                *bun.DB
	accountRepository repository.AccountRepository
}

func NewAccountService(db *bun.DB, accountRepository repository.AccountRepository) AccountService {
	return &accountService{
		db:                db,
		accountRepository: accountRepository,
	}
}

func (s *accountService) CreateAccount(requestCtx context.Context, dbExecutor bun.IDB, userID uuid.UUID, accountType model.AccountType) error {
	if dbExecutor == nil {
		dbExecutor = s.db
	}

	account := &model.Account{
		ID:     s.generateAccountID(),
		UserID: userID,
		Type:   accountType,
	}

	return s.accountRepository.CreateAccount(requestCtx, dbExecutor, account)
}

func (s *accountService) generateAccountID() int64 {
	min := int64(1000000000)      // 10 digits
	max := int64(999999999999999) // 15 digits

	// calculate the range
	rangeNum := big.NewInt(max - min + 1)

	// generate random number in [0, rangeNum)
	n, _ := rand.Int(rand.Reader, rangeNum)

	// shift to the desired range
	accountID := n.Int64() + min
	return accountID
}
