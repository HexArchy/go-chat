package refreshtoken

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/services/auth"
)

type AuthService interface {
	Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
}

type Deps struct {
	AuthService AuthService
}
