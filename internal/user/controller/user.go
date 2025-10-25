package controller

import (
	"net/http"

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
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// transform to DTO and return response
	userDto := types.TransformToGetMeDto(user)
	server.SendSuccessResponse(ginCtx, http.StatusOK, types.GetMeResponse{
		Data: *userDto,
	})
}

func (c *userController) UpdateUser(ginCtx *gin.Context) {
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

	/*
		Note: Currently, username is marked as required in UpdateUserRequest validation
		because it's the only field supported for updates

		In the future, when additional fields are added for updates,,
		the validation should be updated to make all fields optional, and the logic below
		should be modified to dynamically populate UserUpdateOptions only with fields
		that are present in the request (allowing partial updates as expected in PATCH endpoints)
	*/
	var req types.UpdateUserRequest
	isSuccess := server.BindAndValidateIncomingRequestBody(ginCtx, &req)
	if !isSuccess {
		return
	}

	updateOptions := types.UserUpdateOptions{
		Username: &req.Data.Username,
	}

	updatedUser, err := c.userService.UpdateUser(requestCtx, nil, userID, updateOptions)
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// transform to DTO and return response
	userDto := types.TransformToUpdateUserDto(updatedUser)
	server.SendSuccessResponse(ginCtx, http.StatusOK, types.UpdateUserResponse{
		Data: *userDto,
	})
}

func (c *userController) UpdatePassword(ginCtx *gin.Context) {
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

	var req types.UpdatePasswordRequest
	isSuccess := server.BindAndValidateIncomingRequestBody(ginCtx, &req)
	if !isSuccess {
		return
	}

	err := c.userService.UpdatePassword(requestCtx, nil, userID, req.Data.CurrentPassword, req.Data.NewPassword)
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	server.SendSuccessResponse(ginCtx, http.StatusOK, types.UpdatePasswordResponse{
		Success: true,
	})
}
