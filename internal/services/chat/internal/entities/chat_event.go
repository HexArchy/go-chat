package entities

import (
	"time"

	"github.com/google/uuid"
)

type ChatEvent struct {
	RoomID    uuid.UUID
	UserID    uuid.UUID
	Type      EventType
	Message   *Message
	Timestamp time.Time
}
