package disconnect

import (
	"context"

	"github.com/google/uuid"
)

type ChatService interface {
	Disconnect(ctx context.Context, roomID, userID uuid.UUID) error
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (userID uuid.UUID, err error)
}

type Deps struct {
	ChatService ChatService
	AuthService AuthService
}
