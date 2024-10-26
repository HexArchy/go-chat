package tokenmanager

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
)

type AuthClient interface {
	ValidateToken(ctx context.Context, token string) (*entities.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}
