package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/server"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	"github.com/skamranahmed/go-bank/internal/authentication/dto"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	"github.com/uptrace/bun"
)

type authenticationController struct {
	db                    *bun.DB
	authenticationService authenticationService.AuthenticationService
	userService           userService.UserService
	accountService        accountService.AccountService
}

func newAuthenticationController(dependency Dependency) AuthenticationController {
	return &authenticationController{
		db:                    dependency.Db,
		authenticationService: dependency.AuthenticationService,
		userService:           dependency.UserService,
		accountService:        dependency.AccountService,
	}
}

func (c *authenticationController) SignUp(ginCtx *gin.Context) {
	requestCtx := ginCtx.Request.Context()

	var payload dto.SignUpRequest
	isSuccess := server.BindAndValidateIncomingRequestBody(ginCtx, &payload)
	if !isSuccess {
		return
	}

	var accessToken string

	err := c.db.RunInTx(requestCtx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	}, func(ctx context.Context, tx bun.Tx) error {
		// create user record
		userDto, err := c.userService.CreateUser(requestCtx, tx, payload.Data.Email, payload.Data.Password, payload.Data.Username)
		if err != nil {
			return err
		}

		// create an account for user
		err = c.accountService.CreateAccount(requestCtx, tx, userDto.ID)
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
		accessToken, err = c.authenticationService.CreateAccessToken(requestCtx, userDto.ID.String())
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

	server.SendSuccessResponse(ginCtx, http.StatusOK, dto.SignUpResponse{
		AccessToken: accessToken,
	})
}
