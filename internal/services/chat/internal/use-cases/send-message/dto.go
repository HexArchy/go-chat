package sendmessage

import (
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

// MessageInput represents the input data for sending a message.
type MessageInput struct {
	RoomID  uuid.UUID
	UserID  uuid.UUID
	Content string
	Event   *entities.Event
}
