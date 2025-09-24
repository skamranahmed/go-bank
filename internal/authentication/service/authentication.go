package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/cmd/server"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/pkg/cache"
	"github.com/skamranahmed/go-bank/pkg/logger"
	"github.com/uptrace/bun"
)

type authenticationService struct {
	db          *bun.DB
	cacheClient cache.CacheClient
}

func NewAuthenticationService(db *bun.DB, cacheClient cache.CacheClient) AuthenticationService {
	return &authenticationService{
		db:          db,
		cacheClient: cacheClient,
	}
}

func (s *authenticationService) CreateAccessToken(requestCtx context.Context, userID string) (string, error) {
	authConfig := config.GetAuthConfig()

	issuedAt := time.Now().UTC().Unix()
	accessTokenExpiryTTL := time.Duration(authConfig.AccessTokenExpiryDurationInSeconds) * time.Second
	accessTokenExpiresAt := time.Now().Add(accessTokenExpiryTTL).Unix()

	accessTokenPayload := &AccessTokenPayload{
		TokenID:   uuid.NewString(),
		UserID:    userID,
		IssuedAt:  issuedAt,
		ExpiresAt: accessTokenExpiresAt,
	}
	accessToken, err := s.createToken(requestCtx, accessTokenPayload, authConfig.AccessTokenSecretSigningKey)
	if err != nil {
		return "", err
	}

	/*
		Store the access token key in the cache without any value
		The key exists only to:
			1. Check existence (e.g. to validate if the token is active)
			2. Perform deletion on logout
		The actual token data is not stored here and there is no reason to because the token data is contained within the JWT itself
	*/
	accessTokenCacheKey := fmt.Sprintf("auth:access_token_id:%v:user_id:%v", accessTokenPayload.TokenID, userID)
	accessTokenCacheValue := ""
	err = s.cacheClient.SetWithTTL(requestCtx, accessTokenCacheKey, accessTokenCacheValue, accessTokenExpiryTTL)
	if err != nil {
		logger.Error(requestCtx, "Failed to cache access token, error: %+v", err)
		return "", &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to generate access token. Please try again later.",
		}
	}

	return accessToken, nil
}

func (s *authenticationService) VerifyAccessToken(requestCtx context.Context, tokenString string) (*AccessTokenPayload, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %+v", token.Header["alg"])
		}

		if token.Header["alg"] != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("Unexpected signing algorithm: %+v", token.Header["alg"])
		}

		secret := config.GetAuthConfig().AccessTokenSecretSigningKey
		return []byte(secret), nil
	}

	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse token: %+v", err)
	}

	if !token.Valid {
		return nil, errors.New("Invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid token claims")
	}

	issuedAt, ok := claims["issued_at"].(float64)
	if !ok {
		return nil, errors.New("Invalid issued at time")
	}

	expiresAt, ok := claims["expires_at"].(float64)
	if !ok {
		return nil, errors.New("Invalid expiration time")
	}
	if time.Now().Unix() > int64(expiresAt) {
		return nil, errors.New("Token has expired")
	}

	tokenID, ok := claims["token_id"].(string)
	if !ok {
		return nil, errors.New("Invalid token ID")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("Invalid user ID")
	}

	accessTokenCacheKey := fmt.Sprintf("auth:access_token_id:%v:user_id:%v", tokenID, userID)
	_, err = s.cacheClient.Get(requestCtx, accessTokenCacheKey)
	if err != nil {
		return nil, fmt.Errorf("Token not found in cache that means it was either revoked or has been expired: %+v", err)
	}

	return &AccessTokenPayload{
		TokenID:   tokenID,
		UserID:    userID,
		IssuedAt:  int64(issuedAt),
		ExpiresAt: int64(expiresAt),
	}, nil
}

func (s *authenticationService) createToken(requestCtx context.Context, payload any, secretSigningKey string) (string, error) {
	claims := jwt.MapClaims{}

	// keeping it extendible, in case I plan to introduce a refresh token as well
	switch p := payload.(type) {
	case *AccessTokenPayload:
		claims["token_id"] = p.TokenID
		claims["user_id"] = p.UserID
		claims["issued_at"] = p.IssuedAt
		claims["expires_at"] = p.ExpiresAt
	default:
		logger.Error(requestCtx, "Unsupported payload type passed for token creation: %+v", payload)
		return "", &server.ApiError{
			HttpStatusCode: http.StatusInternalServerError,
			Message:        "Unable to generate access token. Please try again later.",
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretSigningKey))
}

type AccessTokenPayload struct {
	TokenID   string `json:"token_id"`
	UserID    string `json:"user_id"`
	IssuedAt  int64  `json:"issued_at"`
	ExpiresAt int64  `json:"expires_at"`
}
