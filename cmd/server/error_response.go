package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApiError struct {
	HttpStatusCode int
	Message        string
}

func (e *ApiError) Error() string {
	return e.Message
}

func SendErrorResponse(ginCtx *gin.Context, err error) {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		sendErrorResponse(ginCtx, apiErr.HttpStatusCode, apiErr.Message)
		return
	}
	sendErrorResponse(ginCtx, http.StatusInternalServerError, err.Error())
}

func sendErrorResponse(ginCtx *gin.Context, httpStatusCode int, customMessage any) {
	var details any
	if msg, ok := customMessage.(string); ok {
		// if customMessage is a string, it should be wrapped inside the "message" key for consistent error format
		details = gin.H{"message": msg}
	} else {
		details = customMessage
	}

	ginCtx.JSON(httpStatusCode, gin.H{
		"error": gin.H{
			"status_code": httpStatusCode,
			"details":     details,
		},
	})
}
