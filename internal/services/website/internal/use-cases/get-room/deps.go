package getroom

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
)

type RoomService interface {
	GetRoom(ctx context.Context, roomID uuid.UUID) (*entities.Room, error)
}

type Deps struct {
	RoomService RoomService
}
