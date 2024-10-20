package getownerrooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
)

type RoomService interface {
	GetOwnerRooms(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error)
}

type Deps struct {
	RoomService RoomService
}
