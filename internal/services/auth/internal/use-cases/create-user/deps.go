package createuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type UserService interface {
	RegisterUser(ctx context.Context, user *entities.User) error
}

type Deps struct {
	UserService UserService
}
