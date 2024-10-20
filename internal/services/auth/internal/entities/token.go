package entities

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
