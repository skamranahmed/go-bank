package controller

import "github.com/gin-gonic/gin"

type UserController interface {
	GetMe(ginCtx *gin.Context)
}
