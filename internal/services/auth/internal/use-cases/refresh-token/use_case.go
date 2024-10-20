package refreshtoken

import (
	"context"

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

func (uc *UseCase) Execute(ctx context.Context, refreshToken string) (string, string, error) {
	newAccessToken, newRefreshToken, err := uc.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to refresh token")
	}

	return newAccessToken, newRefreshToken, nil
}
