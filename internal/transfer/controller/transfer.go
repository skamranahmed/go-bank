package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/middleware"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/internal/account/model"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	accountTypes "github.com/skamranahmed/go-bank/internal/account/types"
	transferService "github.com/skamranahmed/go-bank/internal/transfer/service"
	"github.com/skamranahmed/go-bank/internal/transfer/types"
	"github.com/skamranahmed/go-bank/pkg/database"
	"github.com/uptrace/bun"
)

type transferController struct {
	db              *bun.DB
	transferService transferService.TransferService
	accountService  accountService.AccountService
}

func newTransferController(dependency Dependency) TransferController {
	return &transferController{
		db:              dependency.Db,
		transferService: dependency.TransferService,
		accountService:  dependency.AccountService,
	}
}

func (c *transferController) PerformInternalTransfer(ginCtx *gin.Context) {
	requestCtx := ginCtx.Request.Context()

	// extract user ID from the request context
	userID, ok := requestCtx.Value(middleware.ContextUserIDKey).(string)
	if !ok || userID == "" {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusUnauthorized,
			Message:        "User not authenticated",
		})
		return
	}

	var payload types.InternalTransferRequest
	isSuccess := server.BindAndValidateIncomingRequestBody(ginCtx, &payload)
	if !isSuccess {
		return
	}

	// validate that from and to account ids are different
	// because transferring to the same account doesn't make sense
	if payload.Data.FromAccountID == payload.Data.ToAccountID {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusBadRequest,
			Message:        "Sender and recipient account ids must be different",
		})
		return
	}

	// existence check for the sender account
	fromAccount, err := c.accountService.GetAccount(requestCtx, nil, accountTypes.AccountQueryOptions{
		AccountID: &payload.Data.FromAccountID,
		Columns:   []string{"user_id"},
	})
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// authorization check that sender account belongs to the authenticated user
	if fromAccount.UserID.String() != userID {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusForbidden,
			Message:        "You are not authorized to perform transfer from this account",
		})
		return
	}

	// existence check for the receiver account
	_, err = c.accountService.GetAccount(requestCtx, nil, accountTypes.AccountQueryOptions{
		AccountID: &payload.Data.ToAccountID,
		Columns:   []string{"id"},
	})
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	var senderAccountTransaction *model.Transaction
	err = database.RunInTransaction(requestCtx, "createInternalTransfer", c.db, nil, func(txCtx context.Context, tx bun.Tx) error {
		senderAccountTransaction, err = c.transferService.CreateInternalTransfer(
			txCtx,
			tx,
			fromAccount.UserID,
			payload.Data.FromAccountID,
			payload.Data.ToAccountID,
			*payload.Data.Amount,
		)
		return err
	})
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// transform to DTO and return response
	transactionDto := accountTypes.TransformToTransactionDto(senderAccountTransaction)
	server.SendSuccessResponse(ginCtx, http.StatusOK, types.InternalTransferResponse{
		Data: types.InternalTransferResponseData{
			Transaction: *transactionDto,
		},
	})
}
