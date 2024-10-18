package auth

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthStorage interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, id string) error
	StoreRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, userID, token string) (bool, error)
	RevokeRefreshToken(ctx context.Context, userID, token string) error
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
}
