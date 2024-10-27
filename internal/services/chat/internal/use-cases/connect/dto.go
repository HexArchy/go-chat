package connect

import (
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

// ConnectInput represents the input data for the connect use case.
type ConnectInput struct {
	RoomID     uuid.UUID
	UserID     uuid.UUID
	Connection entities.Connection
	Event      *entities.Event
}
