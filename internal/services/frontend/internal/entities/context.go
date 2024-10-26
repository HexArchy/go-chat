package entities

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type ContextKey string

const (
	ContextKeyUserID       ContextKey = "user_id"
	ContextKeyAccessToken  ContextKey = "accessToken"
	ContextKeyRefreshToken ContextKey = "refreshToken"
)

// GetUserIDFromContext retrieves the user ID from the context.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID); ok {
		return userID, nil
	}
	return uuid.Nil, errors.New("user ID not found in context")
}

// GetAccessTokenFromContext извлекает токен доступа из контекста.
func GetAccessTokenFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value(ContextKeyAccessToken).(string)
	if !ok || token == "" {
		return "", errors.New("access token not found in context")
	}
	return token, nil
}
