package searchrooms

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

func (uc *UseCase) Execute(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error) {
	rooms, err := uc.roomService.SearchRooms(ctx, name, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search rooms")
	}
	return rooms, nil
}
