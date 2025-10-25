package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/server"
	accountModel "github.com/skamranahmed/go-bank/internal/account/model"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	"github.com/skamranahmed/go-bank/internal/authentication/dto"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	userTasks "github.com/skamranahmed/go-bank/internal/user/tasks"
	"github.com/skamranahmed/go-bank/internal/user/types"
	"github.com/skamranahmed/go-bank/pkg/database"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
	"github.com/uptrace/bun"
)

type authenticationController struct {
	db                    *bun.DB
	authenticationService authenticationService.AuthenticationService
	userService           userService.UserService
	accountService        accountService.AccountService
	taskEnqueuer          tasksHelper.TaskEnqueuer
}

func newAuthenticationController(dependency Dependency) AuthenticationController {
	return &authenticationController{
		db:                    dependency.Db,
		authenticationService: dependency.AuthenticationService,
		userService:           dependency.UserService,
		accountService:        dependency.AccountService,
		taskEnqueuer:          dependency.TaskEnqueuer,
	}
}

func (c *authenticationController) SignUp(ginCtx *gin.Context) {
	requestCtx := ginCtx.Request.Context()

	var payload dto.SignUpRequest
	isSuccess := server.BindAndValidateIncomingRequestBody(ginCtx, &payload)
	if !isSuccess {
		return
	}

	var userID, accessToken string
	err := database.RunInTransaction(requestCtx, "signUpTx", c.db, nil, func(txCtx context.Context, tx bun.Tx) error {
		// create user record
		userDto, err := c.userService.CreateUser(txCtx, tx, payload.Data.Email, payload.Data.Password, payload.Data.Username)
		if err != nil {
			return err
		}

		userID = userDto.ID.String()

		// create an account for user
		// currently the API doesn't provide the option to the user to choose the account type during user registration
		// default account type is SAVINGS_ACCOUNT
		err = c.accountService.CreateAccount(txCtx, tx, userDto.ID, accountModel.SavingsAccount)
		if err != nil {
			return err
		}

		/*
			The access token creation is included within the database transaction
			even though it interacts with Redis (not PostgreSQL)

			This ensures atomicity:
			if the user creation in the database succeeds but the token creation fails,
			the new user record that is created is rolled back

			This way, we avoid leaving orphaned users without a valid access token
		*/
		accessToken, err = c.authenticationService.CreateAccessToken(txCtx, userDto.ID.String())
		if err != nil {
			return err
		}

		/*
			If all operations succeed but the final database commit fails,
			the PostgreSQL changes will be rolled back

			The access token in the cache may still exist, but it would be harmless
			since it won't be associated with any user and will expire automatically due to its TTL
		*/
		return nil
	})

	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	// send welcome email task
	err = c.taskEnqueuer.Enqueue(requestCtx, userTasks.NewSendWelcomeEmailTask(userID), nil, nil)
	if err != nil {
		logger.Error(requestCtx, "Unable to enqueue SendWelcomeEmailTask for userID: %s, error: %+v", userID, err)
	}

	server.SendSuccessResponse(ginCtx, http.StatusCreated, dto.SignUpResponse{
		AccessToken: accessToken,
	})
}

func (c *authenticationController) Login(ginCtx *gin.Context) {
	requestCtx := ginCtx.Request.Context()

	var payload dto.LoginRequest
	isSuccess := server.BindAndValidateIncomingRequestBody(ginCtx, &payload)
	if !isSuccess {
		return
	}

	userQueryOptions := types.UserQueryOptions{
		Username: &payload.Data.Username,
	}
	user, err := c.userService.GetUser(requestCtx, nil, userQueryOptions)
	if err != nil {
		var apiError *server.ApiError
		if errors.As(err, &apiError) && apiError.HttpStatusCode == http.StatusNotFound {
			// if user not found, return a generic authentication error for better security
			// we should not reveal whether the user doesn't exist or the password was incorrect
			server.SendErrorResponse(ginCtx, &server.ApiError{
				HttpStatusCode: http.StatusUnauthorized,
				Message:        "Invalid username or password",
			})
			return
		}
		server.SendErrorResponse(ginCtx, err)
		return
	}

	doesPasswordMatch, err := argon2id.ComparePasswordAndHash(payload.Data.Password, user.Password)
	if err != nil {
		logger.Error(requestCtx, "Error comparing password hash, error: %+v", err)
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to process your request. Please try again later.",
		})
		return
	}

	if !doesPasswordMatch {
		server.SendErrorResponse(ginCtx, &server.ApiError{
			HttpStatusCode: http.StatusUnauthorized,
			Message:        "Invalid username or password",
		})
		return
	}

	accessToken, err := c.authenticationService.CreateAccessToken(requestCtx, user.ID.String())
	if err != nil {
		server.SendErrorResponse(ginCtx, err)
		return
	}

	server.SendSuccessResponse(ginCtx, http.StatusOK, dto.LoginResponse{
		AccessToken: accessToken,
	})
}
