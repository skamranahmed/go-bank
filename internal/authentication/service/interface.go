package service

import "context"

type AuthenticationService interface {
	CreateAccessToken(requestCtx context.Context, userID string) (string, error)
	VerifyAccessToken(requestCtx context.Context, tokenString string) (*AccessTokenPayload, error)
}
