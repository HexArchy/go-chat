package login

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

func (uc *UseCase) Execute(ctx context.Context, email, password string) (*LoginResult, error) {
	tokenPair, err := uc.authService.Login(ctx, email, password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}

	return &LoginResult{
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
	}, nil
}
