package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/cmd/middleware"
	"github.com/skamranahmed/go-bank/cmd/server"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	"github.com/skamranahmed/go-bank/internal/account/types"
)

type accountController struct {
	accountService accountService.AccountService
}

func newAccountController(dependency Dependency) AccountController {
	return &accountController{
		accountService: dependency.AccountService,
	}
}

func (c *accountController) GetAccounts(ginCtx *gin.Context) {
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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusBadRequest,
			Message:        "Invalid user ID",
		})
		return
	}

	accounts, err := c.accountService.GetAccountsByUserID(requestCtx, nil, userUUID)
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// transform to DTO and return response
	accountDtos := types.TransformToAccountDtoList(accounts)
	server.SendSuccessResponse(ginCtx, http.StatusOK, types.GetAccountsResponse{
		Data: accountDtos,
	})
}

func (c *accountController) GetAccountByID(ginCtx *gin.Context) {
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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusBadRequest,
			Message:        "Invalid user ID",
		})
		return
	}

	// extract account ID from URL parameter
	accountIDParam := ginCtx.Param("account_id")
	accountID, err := strconv.ParseInt(accountIDParam, 10, 64)
	if err != nil {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusBadRequest,
			Message:        "Invalid account ID",
		})
		return
	}

	account, err := c.accountService.GetAccount(requestCtx, nil, types.AccountQueryOptions{
		AccountID: &accountID,
	})
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// authorization check: verify account belongs to authenticated user
	if account.UserID != userUUID {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusForbidden,
			Message:        "You do not have permission to access this account",
		})
		return
	}

	// transform to DTO and return response
	accountDto := types.TransformToAccountDto(account)
	server.SendSuccessResponse(ginCtx, http.StatusOK, types.GetAccountByIDResponse{
		Data: *accountDto,
	})
}
