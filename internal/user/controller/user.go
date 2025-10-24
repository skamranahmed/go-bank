package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/middleware"
	"github.com/skamranahmed/go-bank/cmd/server"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	"github.com/skamranahmed/go-bank/internal/user/types"
)

type userController struct {
	userService userService.UserService
}

func newUserController(dependency Dependency) UserController {
	return &userController{
		userService: dependency.UserService,
	}
}

func (c *userController) GetMe(ginCtx *gin.Context) {
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

	userQueryOptions := types.UserQueryOptions{
		ID: &userID,
	}
	user, err := c.userService.GetUser(requestCtx, nil, userQueryOptions)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			server.SendErrorResponse(ginCtx, &server.ApiError{
				HttpStatusCode: http.StatusNotFound,
				Message:        "User not found",
			})
			return
		}
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// transform to DTO and return response
	userDto := types.TransformToGetMeDto(user)
	server.SendSuccessResponse(ginCtx, http.StatusOK, types.GetMeResponse{
		Data: *userDto,
	})
}
