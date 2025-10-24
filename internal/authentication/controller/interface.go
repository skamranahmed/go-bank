package controller

import "github.com/gin-gonic/gin"

type AuthenticationController interface {
	SignUp(ginCtx *gin.Context)
	Login(ginCtx *gin.Context)
}
