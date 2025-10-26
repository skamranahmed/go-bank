package controller

import "github.com/gin-gonic/gin"

type TransferController interface {
	PerformInternalTransfer(ginCtx *gin.Context)
}
