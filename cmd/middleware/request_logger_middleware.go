package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		start := time.Now()

		correlationID := xid.New().String()

		// attach correlationID to request context
		ctx := context.WithValue(ginCtx.Request.Context(), "correlation_id", correlationID)
		ginCtx.Request = ginCtx.Request.WithContext(ctx)
		requestCtx := ginCtx.Request.Context()

		defer func() {
			logMessage := "Request processed"

			requestProcessingDuration := time.Since(start)
			responseTimeInPlainMs, humanReadableResponseTime := getResponseTime(requestProcessingDuration)

			// check whether there was any panic recovery
			r := recover()
			if r != nil {
				logMessage = "Request processed with panic recovery"
				errMsg := fmt.Sprintf("%v", r)
				stack := string(debug.Stack())
				logger.Error(requestCtx, "Panic recovered, errMsg: %+v, stackTrace: %+v", errMsg, stack)
				ginCtx.Writer.WriteHeader(http.StatusInternalServerError)
			}

			// collect and log useful fields
			logger.InfoFields(logMessage, map[string]any{
				"request_method":          ginCtx.Request.Method,
				"request_path":            ginCtx.Request.URL.Path,
				"request_query_params":    ginCtx.Request.URL.RawQuery,
				"request_referer":         ginCtx.Request.Referer(),
				"client_ip":               ginCtx.ClientIP(),
				"user_agent":              ginCtx.Request.UserAgent(),
				"response_status":         ginCtx.Writer.Status(),
				"response_length":         ginCtx.Writer.Size(),
				"response_time_formatted": humanReadableResponseTime,
				"response_time_ms":        responseTimeInPlainMs,
				"correlation_id":          correlationID,
			})
		}()

		// process the request
		ginCtx.Next()

	}
}

// returns numeric milliseconds (float64) and human-readable formatted string
func getResponseTime(d time.Duration) (ms float64, formatted string) {
	ms = float64(d.Microseconds()) / 1000 // milliseconds with fractions

	switch {
	case ms < 1000:
		formatted = fmt.Sprintf("%.3fms", ms)
	case d.Seconds() < 60:
		formatted = fmt.Sprintf("%.3fs", d.Seconds())
	default:
		formatted = fmt.Sprintf("%.3fmin", d.Minutes())
	}

	return
}
