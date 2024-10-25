package entities

import (
	"time"

	"github.com/google/uuid"
)

type ChatParticipant struct {
	RoomID   uuid.UUID
	UserID   uuid.UUID
	JoinedAt time.Time
}
