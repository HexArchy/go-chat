package createuser

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

func (uc *UseCase) Execute(ctx context.Context, user *entities.User) error {
	if err := user.ValidateEmail(); err != nil {
		return errors.Wrap(err, "invalid email format")
	}

	if err := user.ValidatePassword(); err != nil {
		return errors.Wrap(err, "invalid password")
	}

	if err := uc.userService.RegisterUser(ctx, user); err != nil {
		return errors.Wrap(err, "failed to register user")
	}

	return nil
}
