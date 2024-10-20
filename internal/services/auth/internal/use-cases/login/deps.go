package login

import (
	"context"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
}

type Deps struct {
	AuthService AuthService
}
