package entities

import (
	"errors"
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

// ErrRoomNotFound is used when a room could not be found.
var ErrRoomNotFound = errors.New("room not found")

// ErrRoomAlreadyExists is used when a room with the same name already exists.
var ErrRoomAlreadyExists = errors.New("room already exists")

// ErrRoomDeleteForbidden is used when an unauthorized user tries to delete a room.
var ErrRoomDeleteForbidden = errors.New("you are not allowed to delete this room")
