package server

import "github.com/gin-gonic/gin"

func SendSuccessResponse(ginCtx *gin.Context, httpStatusCode int, response any) {
	if response == nil {
		ginCtx.Status(httpStatusCode)
		return
	}
	ginCtx.JSON(httpStatusCode, response)
}
