package entities

import (
	"time"

	"github.com/google/uuid"
)

// Message представляет сообщение в чате
type Message struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	UserID    uuid.UUID
	Content   string
	CreatedAt time.Time
}
