package entities

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID
	Name      string
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
