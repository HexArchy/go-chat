package refreshtoken

import (
	"context"
)

type AuthService interface {
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
}

type Deps struct {
	AuthService AuthService
}
