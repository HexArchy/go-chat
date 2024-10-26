package logout

import (
	"context"

	"github.com/google/uuid"
)

type AuthService interface {
	Revoke(ctx context.Context, userID uuid.UUID) error
}

type Deps struct {
	AuthService AuthService
}
