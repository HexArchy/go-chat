package validatetoken

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthService interface {
	ValidateToken(ctx context.Context, tokenString string) (*entities.User, error)
}

type Deps struct {
	AuthService AuthService
}
