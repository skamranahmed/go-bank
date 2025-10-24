package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/middleware"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
	userService "github.com/skamranahmed/go-bank/internal/user/service"
)

type Dependency struct {
	AuthenticationService authenticationService.AuthenticationService
	UserService           userService.UserService
}

func Register(router *gin.Engine, dependency Dependency) {
	userController := newUserController(dependency)
	router.GET("/v1/me", middleware.AuthMiddleware(middleware.AuthMandatory, dependency.AuthenticationService), userController.GetMe)
}
