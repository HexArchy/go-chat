package getmessages

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

// ChatService defines the interface for chat operations.
type ChatService interface {
	GetRoomMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error)
}

// Deps holds the dependencies for the get messages use case.
type Deps struct {
	ChatService ChatService
}
