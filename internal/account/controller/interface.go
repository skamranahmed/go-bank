package controller

import "github.com/gin-gonic/gin"

type AccountController interface {
	GetAccounts(ginCtx *gin.Context)
	GetAccountByID(ginCtx *gin.Context)
}
