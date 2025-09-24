package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/cmd/worker"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	"github.com/skamranahmed/go-bank/internal/authentication/dto"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	userTasks "github.com/skamranahmed/go-bank/internal/user/tasks"
	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/uptrace/bun"
)

type authenticationController struct {
	db                    *bun.DB
	authenticationService authenticationService.AuthenticationService
	userService           userService.UserService
	accountService        accountService.AccountService
	asynqService          *asynq.Client
}

func newAuthenticationController(dependency Dependency) AuthenticationController {
	return &authenticationController{
		db:                    dependency.Db,
		authenticationService: dependency.AuthenticationService,
		userService:           dependency.UserService,
		accountService:        dependency.AccountService,
		asynqService:          dependency.AsynqService,
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
	err := c.db.RunInTx(requestCtx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	}, func(ctx context.Context, tx bun.Tx) error {
		// create user record
		userDto, err := c.userService.CreateUser(requestCtx, tx, payload.Data.Email, payload.Data.Password, payload.Data.Username)
		if err != nil {
			return err
		}

		userID = userDto.ID.String()

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

	// send welcome email task
	task, err := userTasks.NewSendWelcomeEmailTask(requestCtx, userID)
	if err != nil {
		logger.Error(requestCtx, "Unable to create SendWelcomeEmailTask, error: %+v", err)
	}

	taskInfo, err := c.asynqService.Enqueue(task, asynq.Queue(worker.DefaultQueue))
	if err != nil {
		logger.Error(requestCtx, "Unable to enqueue SendWelcomeEmailTask, error: %+v", err)
	} else {
		logger.Info(requestCtx, "Enqueued SendWelcomeEmailTask for userID: %+v, taskID: %+v, queue: %+v", userID, taskInfo.ID, taskInfo.Queue)
	}

	server.SendSuccessResponse(ginCtx, http.StatusCreated, dto.SignUpResponse{
		AccessToken: accessToken,
	})
}
