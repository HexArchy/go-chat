package chat

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

type MessageStorage interface {
	CreateMessage(ctx context.Context, message *entities.Message) error
	GetMessage(ctx context.Context, messageID uuid.UUID) (*entities.Message, error)
	GetRoomMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error)

	AddParticipant(ctx context.Context, roomID, userID uuid.UUID) error
	RemoveParticipant(ctx context.Context, roomID, userID uuid.UUID) error
	GetRoomParticipants(ctx context.Context, roomID uuid.UUID) ([]*entities.ChatParticipant, error)
	IsParticipant(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
}

type WebsiteClient interface {
	GetRoom(ctx context.Context, roomID uuid.UUID) (*entities.Room, error)
	IsRoomOwner(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
	RoomExists(ctx context.Context, roomID uuid.UUID) (bool, error)
}

type Deps struct {
	MessageStorage MessageStorage
	WebsiteClient  WebsiteClient
}
