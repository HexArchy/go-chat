package user

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
}

type Deps struct {
	UserStorage UserStorage
}
