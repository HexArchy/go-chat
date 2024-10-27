package getroomparticipants

import (
	"context"

	"github.com/google/uuid"
)

// Storage defines the interface for chat data persistence.
type Storage interface {
	GetRoomParticipants(ctx context.Context, roomID uuid.UUID) ([]uuid.UUID, error)
}

// WebsiteService defines the interface for room validation.
type WebsiteService interface {
	RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

// Deps holds the dependencies for the get room participants use case.
type Deps struct {
	Storage        Storage
	WebsiteService WebsiteService
}
