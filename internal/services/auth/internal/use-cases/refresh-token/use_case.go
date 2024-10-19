package refreshtoken

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

func (uc *UseCase) Execute(ctx context.Context, refreshToken string) (entities.AuthenticatedUser, error) {
	// Validate input.
	if refreshToken == "" {
		return entities.AuthenticatedUser{}, errors.Wrap(entities.ErrInvalidInput, "refresh token is required")
	}

	// Refresh tokens.
	authenticatedUser, err := uc.authService.RefreshTokens(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidRefreshToken) {
			return entities.AuthenticatedUser{}, entities.ErrInvalidRefreshToken
		}
		return entities.AuthenticatedUser{}, errors.Wrap(err, "failed to refresh tokens")
	}

	return authenticatedUser, nil
}
