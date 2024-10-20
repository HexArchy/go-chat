package deleteuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
)

type UserService interface {
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	CheckPermission(ctx context.Context, userID uuid.UUID, permission entities.Permission) (bool, error)
}

type Deps struct {
	UserService UserService
}
