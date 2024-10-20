package deleteuser

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

func (uc *UseCase) Execute(ctx context.Context, requesterID, targetUserID uuid.UUID) error {
	hasPermission, err := uc.userService.CheckPermission(ctx, requesterID, "admin")
	if err != nil {
		return errors.Wrap(err, "failed to check permissions")
	}

	if !hasPermission && requesterID != targetUserID {
		return entities.ErrPermissionDenied
	}

	if err := uc.userService.DeleteUser(ctx, targetUserID); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	return nil
}
