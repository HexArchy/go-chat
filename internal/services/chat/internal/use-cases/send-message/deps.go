package sendmessage

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

// ChatService defines the interface for chat operations.
type ChatService interface {
	HandleMessage(ctx context.Context, roomID, userID uuid.UUID, content string, event *entities.Event) error
}

// Deps holds the dependencies for the send message use case.
type Deps struct {
	ChatService ChatService
}
