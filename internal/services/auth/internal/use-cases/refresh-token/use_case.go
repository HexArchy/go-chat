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

func (uc *UseCase) Execute(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	tokenPair, err := uc.authService.Refresh(ctx, refreshToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh token")
	}

	return &RefreshResult{
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
	}, nil
}
