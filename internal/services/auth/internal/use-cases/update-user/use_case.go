package updateuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"

	"github.com/google/uuid"
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

func (uc *UseCase) Execute(ctx context.Context, requesterID uuid.UUID, user *entities.User) error {
	hasPermission, err := uc.userService.CheckPermission(ctx, requesterID, "admin")
	if err != nil {
		return errors.Wrap(err, "failed to check permissions")
	}

	if !hasPermission && requesterID != user.ID {
		return entities.ErrPermissionDenied
	}

	if user.Email != "" {
		if err = user.ValidateEmail(); err != nil {
			return errors.Wrap(err, "failed to validate email")
		}
	}

	if user.Password != "" {
		if err = user.ValidatePassword(); err != nil {
			return errors.Wrap(err, "failed to validate password")
		}
	}

	if err := uc.userService.UpdateUser(ctx, user); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	return nil
}
