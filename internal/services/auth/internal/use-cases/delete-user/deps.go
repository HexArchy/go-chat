package deleteuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type AuthService interface {
	RevokeAllRefreshTokens(ctx context.Context, userID string) error
}

type AuthStorage interface {
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	DeleteUser(ctx context.Context, id string) error
}

type RBAC interface {
	CheckPermission(roles []entities.Role, action, resource string) bool
}
