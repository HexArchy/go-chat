package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Email       string
	Username    string
	Phone       string
	Age         int32
	Bio         string
	Permissions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
