package controller

import "github.com/gin-gonic/gin"

type HealthzController interface {
	CheckHealth(ginCtx *gin.Context)
}
