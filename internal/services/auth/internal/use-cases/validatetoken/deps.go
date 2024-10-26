package validatetoken

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthService interface {
	Validate(ctx context.Context, accessToken string) (*entities.User, error)
}

type Deps struct {
	AuthService AuthService
}
