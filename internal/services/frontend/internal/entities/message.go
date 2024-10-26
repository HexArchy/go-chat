package entities

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	UserID    uuid.UUID
	Content   string
	CreatedAt time.Time
}
