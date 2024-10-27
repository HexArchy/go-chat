package disconnect

import (
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

type DisconnectInput struct {
	RoomID uuid.UUID
	UserID uuid.UUID
	Event  *entities.Event
}
