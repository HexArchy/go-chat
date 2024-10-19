package updateuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthService interface {
	ValidateToken(ctx context.Context, tokenString, tokenType string) (*entities.AuthenticatedUser, error)
}

type AuthStorage interface {
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
}

type RBAC interface {
	CheckPermission(roles []entities.Role, action, resource string) bool
}
