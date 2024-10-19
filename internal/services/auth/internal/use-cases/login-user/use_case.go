package loginuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/pkg/errors"
)

type UseCase struct {
	authService AuthService
}

func New(authService AuthService) *UseCase {
	return &UseCase{
		authService: authService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, email, password string) (entities.AuthenticatedUser, error) {
	// Validate input.
	if email == "" || password == "" {
		return entities.AuthenticatedUser{}, errors.Wrap(entities.ErrInvalidInput, "email and password are required")
	}

	// Authenticate user.
	authenticatedUser, err := uc.authService.AuthenticateUser(ctx, email, password)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidCredantials) {
			return entities.AuthenticatedUser{}, entities.ErrInvalidCredantials
		}
		return entities.AuthenticatedUser{}, errors.Wrap(err, "failed to authenticate user")
	}

	return authenticatedUser, nil
}
