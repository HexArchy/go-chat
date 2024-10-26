package validatetoken

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
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

func (uc *UseCase) Execute(ctx context.Context, accessToken string) (*entities.User, error) {
	user, err := uc.authService.Validate(ctx, accessToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate token")
	}

	return user, nil
}
