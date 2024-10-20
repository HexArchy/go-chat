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

func (uc *UseCase) Execute(ctx context.Context, email, password string) (string, string, error) {
	accessToken, refreshToken, err := uc.authService.Login(ctx, email, password)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to login")
	}

	return accessToken, refreshToken, nil
}
