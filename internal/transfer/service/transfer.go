package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/cmd/server"
	accountModel "github.com/skamranahmed/go-bank/internal/account/model"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	accountTypes "github.com/skamranahmed/go-bank/internal/account/types"
	"github.com/uptrace/bun"
)

type transferService struct {
	db             *bun.DB
	accountService accountService.AccountService
}

func NewTransferService(db *bun.DB, accountService accountService.AccountService) TransferService {
	return &transferService{
		db:             db,
		accountService: accountService,
	}
}

func (s *transferService) CreateInternalTransfer(
	requestCtx context.Context,
	dbExecutor bun.IDB,
	senderUserID uuid.UUID,
	fromAccountID, toAccountID, transferAmount int64,
) (*accountModel.Transaction, error) {
	if dbExecutor == nil {
		dbExecutor = s.db
	}

	/*
		Prevent deadlocks by establishing a consistent ordering for row locking

		When multiple concurrent transactions involve the same two accounts in different roles:
		1. Transaction A: account1->account2
		2. Transaction B: account2->account1
		we must lock accounts in a deterministic order regardless of sender/receiver roles

		By always locking the account with either the smaller or bigger ID first, we ensure all transactions
		follow the same locking sequence, preventing circular wait conditions that cause deadlocks

		For the implementation here, we will lock accounts in ascending ID order to maintain consistency across all transactions
		Ascending ID order means, the smaller account ID will always be locked first
	*/
	var firstAccountID, secondAccountID int64
	if fromAccountID < toAccountID {
		firstAccountID = fromAccountID
		secondAccountID = toAccountID
	} else {
		firstAccountID = toAccountID
		secondAccountID = fromAccountID
	}

	var err error
	var firstAccount *accountModel.Account
	firstAccount, err = s.accountService.GetAccount(requestCtx, dbExecutor, accountTypes.AccountQueryOptions{
		AccountID: &firstAccountID,
		ForUpdate: true, // lock the row for update
	})
	if err != nil {
		return nil, err
	}

	var secondAccount *accountModel.Account
	secondAccount, err = s.accountService.GetAccount(requestCtx, dbExecutor, accountTypes.AccountQueryOptions{
		AccountID: &secondAccountID,
		ForUpdate: true, // lock the row for update
	})
	if err != nil {
		return nil, err
	}

	// check which of the account is the sender account and check if it has enough balance
	var senderAccount, receiverAccount *accountModel.Account
	if firstAccount.ID == fromAccountID {
		senderAccount = firstAccount
		receiverAccount = secondAccount
	} else {
		senderAccount = secondAccount
		receiverAccount = firstAccount
	}

	if senderAccount.Balance < transferAmount {
		return nil, &server.ApiError{
			HttpStatusCode: http.StatusBadRequest,
			Message:        "You do not have sufficient balance in your account to perform the transfer",
		}
	}

	// update the balance of the sender's account (debit)
	updatedBalanceAfterDebit := senderAccount.Balance - transferAmount
	senderAccount, err = s.accountService.UpdateAccount(requestCtx, dbExecutor, senderAccount.ID, accountTypes.AccountUpdateOptions{
		NewBalance: &updatedBalanceAfterDebit,
	})
	if err != nil {
		return nil, err
	}

	// create transaction record for sender account
	transactionRecordForSenderAccount := &accountModel.Transaction{
		AccountID:    senderAccount.ID,
		Amount:       transferAmount,
		BalanceAfter: senderAccount.Balance,
		Type:         accountModel.Debit, // debit transaction
	}
	transactionRecordForSenderAccount, err = s.accountService.CreateTransactionRecord(requestCtx, dbExecutor, transactionRecordForSenderAccount)
	if err != nil {
		return nil, err
	}

	// update the balance of the receiver's account (credit)
	updatedBalanceAfterCredit := receiverAccount.Balance + transferAmount
	receiverAccount, err = s.accountService.UpdateAccount(requestCtx, dbExecutor, receiverAccount.ID, accountTypes.AccountUpdateOptions{
		NewBalance: &updatedBalanceAfterCredit,
	})
	if err != nil {
		return nil, err
	}

	// create transaction record for receiver account
	transactionRecordForReceiverAccount := &accountModel.Transaction{
		AccountID:    receiverAccount.ID,
		Amount:       transferAmount,
		BalanceAfter: receiverAccount.Balance,
		Type:         accountModel.Credit,
	}
	_, err = s.accountService.CreateTransactionRecord(requestCtx, dbExecutor, transactionRecordForReceiverAccount)
	if err != nil {
		return nil, err
	}

	return transactionRecordForSenderAccount, nil
}
