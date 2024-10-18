package refreshtoken

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthService interface {
	RefreshTokens(ctx context.Context, refreshToken string) (entities.AuthenticatedUser, error)
}
