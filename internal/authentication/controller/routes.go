package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	"github.com/uptrace/bun"
)

type Dependency struct {
	Db                    *bun.DB
	AuthenticationService authenticationService.AuthenticationService
	UserService           userService.UserService
	AccountService        accountService.AccountService
	AsynqService          *asynq.Client
}

func Register(router *gin.Engine, dependency Dependency) {
	authenticationController := newAuthenticationController(dependency)
	router.POST("/v1/sign-up", authenticationController.SignUp)
}
