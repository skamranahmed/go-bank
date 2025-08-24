package controller

import (
	"github.com/gin-gonic/gin"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
	"github.com/uptrace/bun"
)

type Dependency struct {
	Db                    *bun.DB
	AuthenticationService authenticationService.AuthenticationService
	UserService           userService.UserService
}

func Register(router *gin.Engine, dependency Dependency) {
	authenticationController := newAuthenticationController(dependency)
	router.POST("/v1/sign-up", authenticationController.SignUp)
}
