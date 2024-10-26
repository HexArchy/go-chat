package entities

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
