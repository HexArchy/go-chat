package deleteroom

import (
	"context"

	"github.com/google/uuid"
)

type RoomService interface {
	DeleteRoom(ctx context.Context, roomID, ownerID uuid.UUID) error
}

type Deps struct {
	RoomService RoomService
}
