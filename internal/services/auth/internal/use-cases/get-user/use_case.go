package getuser

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

func (uc *UseCase) Execute(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	user, err := uc.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	return user, nil
}
