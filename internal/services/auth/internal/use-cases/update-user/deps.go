package updateuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
)

type UserService interface {
	UpdateUser(ctx context.Context, user *entities.User) error
	CheckPermission(ctx context.Context, userID uuid.UUID, permission entities.Permission) (bool, error)
}

type Deps struct {
	UserService UserService
}
