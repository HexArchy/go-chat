package connect

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

type WebsiteService interface {
	RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type ChatService interface {
	Connect(ctx context.Context, roomID, userID uuid.UUID, conn entities.ChatConnection) error
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (userID uuid.UUID, err error)
}

type Deps struct {
	WebsiteService WebsiteService
	ChatService    ChatService
	AuthService    AuthService
}
