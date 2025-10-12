package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
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

	errObject := gin.H{
		"error": gin.H{
			"status_code": httpStatusCode,
			"details":     details,
		},
	}

	// record the error in the root span of the request
	rootSpan := trace.SpanFromContext(ginCtx.Request.Context())

	// marshal the error object to JSON for nicer formatting
	errJSON, err := json.MarshalIndent(errObject, "", "  ")
	if err != nil {
		// fallback in case JSON marshaling fails
		rootSpan.RecordError(fmt.Errorf("api response: %+v", errObject))
	} else {
		rootSpan.RecordError(fmt.Errorf("api response: %s", string(errJSON)))
	}

	ginCtx.JSON(httpStatusCode, errObject)
}
