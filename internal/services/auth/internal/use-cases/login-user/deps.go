package loginuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthService interface {
	AuthenticateUser(ctx context.Context, email, password string) (entities.AuthenticatedUser, error)
}
