package disconnect

import (
	"context"

	"github.com/google/uuid"
)

// ChatService defines the interface for chat operations.
type ChatService interface {
	Disconnect(ctx context.Context, roomID, userID uuid.UUID) error
}

// Deps holds the dependencies for the disconnect use case.
type Deps struct {
	ChatService ChatService
}
