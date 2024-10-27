package connect

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

// WebsiteService defines the interface for room validation.
type WebsiteService interface {
	RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

// ChatService defines the interface for chat operations.
type ChatService interface {
	Connect(ctx context.Context, roomID, userID uuid.UUID, conn entities.Connection) error
}

// Deps holds the dependencies for the connect use case.
type Deps struct {
	WebsiteService WebsiteService
	ChatService    ChatService
}
