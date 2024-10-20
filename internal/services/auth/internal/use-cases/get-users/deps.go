package getusers

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type UserService interface {
	GetUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
}

type Deps struct {
	UserService UserService
}
