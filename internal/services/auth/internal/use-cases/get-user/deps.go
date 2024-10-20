package getuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
)

type UserService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*entities.User, error)
}

type Deps struct {
	UserService UserService
}
