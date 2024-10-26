package getallrooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/pkg/errors"
)

type UseCase struct {
	roomService RoomService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		roomService: deps.RoomService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, limit, offset int) ([]*entities.Room, error) {
	rooms, err := uc.roomService.GetAllRooms(ctx, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve all rooms")
	}
	return rooms, nil
}
