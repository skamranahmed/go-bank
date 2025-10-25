package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/middleware"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
)

type Dependency struct {
	AuthenticationService authenticationService.AuthenticationService
	AccountService        accountService.AccountService
}

func Register(router *gin.Engine, dependency Dependency) {
	accountController := newAccountController(dependency)
	router.GET("/v1/accounts", middleware.AuthMiddleware(middleware.AuthMandatory, dependency.AuthenticationService), accountController.GetAccounts)
}
