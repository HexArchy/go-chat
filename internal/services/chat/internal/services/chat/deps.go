package chat

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

// Storage defines the interface for chat data persistence.
type Storage interface {
	AddParticipant(ctx context.Context, roomID, userID uuid.UUID) error
	RemoveParticipant(ctx context.Context, roomID, userID uuid.UUID) error
	IsParticipant(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
	SaveMessage(ctx context.Context, message *entities.Message) error
	GetLastMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error)
}

type WebsiteService interface {
	RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type Deps struct {
	Storage Storage
}
