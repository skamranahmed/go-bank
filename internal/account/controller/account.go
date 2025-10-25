package controller

import (
	"net/http"

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
