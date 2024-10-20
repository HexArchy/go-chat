package getusers

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/pkg/errors"
)

type UseCase struct {
	userService UserService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		userService: deps.UserService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	users, err := uc.userService.GetUsers(ctx, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get users")
	}

	return users, nil
}
