package getallrooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
)

type RoomService interface {
	GetAllRooms(ctx context.Context, limit, offset int) ([]*entities.Room, error)
}

type Deps struct {
	RoomService RoomService
}
