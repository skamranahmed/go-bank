package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/middleware"
	accountService "github.com/skamranahmed/go-bank/internal/account/service"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	transferService "github.com/skamranahmed/go-bank/internal/transfer/service"
	"github.com/uptrace/bun"
)

type Dependency struct {
	Db                    *bun.DB
	AuthenticationService authenticationService.AuthenticationService
	AccountService        accountService.AccountService
	TransferService       transferService.TransferService
}

func Register(router *gin.Engine, dependency Dependency) {
	transferController := newTransferController(dependency)
	router.POST("/v1/transfers/internal", middleware.AuthMiddleware(middleware.AuthMandatory, dependency.AuthenticationService), transferController.PerformInternalTransfer)
}
