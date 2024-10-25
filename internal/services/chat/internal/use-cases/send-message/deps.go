package sendmessage

import (
	"context"

	"github.com/google/uuid"
)

type ChatService interface {
	SendMessage(ctx context.Context, roomID, userID uuid.UUID, content string) error
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (userID uuid.UUID, err error)
}

type Deps struct {
	ChatService ChatService
	AuthService AuthService
}
