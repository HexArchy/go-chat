package logout

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UseCase struct {
	authService AuthService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		authService: deps.AuthService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	if err := uc.authService.Revoke(ctx, userID); err != nil {
		return errors.Wrap(err, "failed to logout user")
	}

	return nil
}
