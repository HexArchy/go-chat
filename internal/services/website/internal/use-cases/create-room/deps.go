package createroom

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
)

type RoomService interface {
	CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*entities.Room, error)
}

type Deps struct {
	RoomService RoomService
}
