package login

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/services/auth"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*auth.TokenPair, error)
}

type Deps struct {
	AuthService AuthService
}
