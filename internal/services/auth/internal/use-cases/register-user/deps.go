package registeruser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthService interface {
	AuthenticateUser(ctx context.Context, email, password string) (entities.AuthenticatedUser, error)
	RegisterUser(ctx context.Context, user *entities.User) error
}

type AuthStorage interface {
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
}
