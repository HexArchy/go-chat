package searchrooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
)

type RoomService interface {
	SearchRooms(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error)
}

type Deps struct {
	RoomService RoomService
}
