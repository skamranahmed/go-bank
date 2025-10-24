package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skamranahmed/go-bank/cmd/server"
	authenticationService "github.com/skamranahmed/go-bank/internal/authentication/service"
)

const ContextUserIDKey = "authUserID"

// AuthMode determines if auth is mandatory or optional
type AuthMode int

const (
	AuthMandatory AuthMode = iota
	AuthOptional
)

// AuthMiddleware returns a Gin middleware for mandatory or optional auth
func AuthMiddleware(mode AuthMode, authService authenticationService.AuthenticationService) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		authHeader := ginCtx.GetHeader("Authorization")
		if authHeader == "" {
			if mode == AuthMandatory {
				server.SendErrorResponse(ginCtx, &server.ApiError{
					HttpStatusCode: http.StatusUnauthorized,
					Message:        "Authorization header is missing",
				})
				ginCtx.Abort()
				return
			}
			// optional: no token, just continue
			ginCtx.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			server.SendErrorResponse(ginCtx, &server.ApiError{
				HttpStatusCode: http.StatusUnauthorized,
				Message:        "Invalid Authorization header format",
			})
			ginCtx.Abort()
			return
		}

		tokenString := parts[1]
		payload, err := authService.VerifyAccessToken(ginCtx.Request.Context(), tokenString)
		if err != nil {
			server.SendErrorResponse(ginCtx, &server.ApiError{
				HttpStatusCode: http.StatusUnauthorized,
				Message:        "Invalid or expired token",
			})
			ginCtx.Abort()
			return
		}

		// attach userID to request context
		ctx := context.WithValue(ginCtx.Request.Context(), ContextUserIDKey, payload.UserID)
		ginCtx.Request = ginCtx.Request.WithContext(ctx)
		ginCtx.Next()
	}
}
