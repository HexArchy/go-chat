package getmessages

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

type ChatService interface {
	GetMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error)
}

type WebsiteService interface {
	RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type Deps struct {
	ChatService    ChatService
	WebsiteService WebsiteService
}
