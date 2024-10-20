package rooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
)

type RoomStorage interface {
	CreateRoom(ctx context.Context, room *entities.Room) error
	GetRoomByID(ctx context.Context, roomID uuid.UUID) (*entities.Room, error)
	GetRoomsByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error)
	GetRoomsByName(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error)
	DeleteRoom(ctx context.Context, roomID uuid.UUID) error
}

type Deps struct {
	RoomStorage RoomStorage
}
