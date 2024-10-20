package getroom

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
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

func (uc *UseCase) Execute(ctx context.Context, roomID uuid.UUID) (*entities.Room, error) {
	room, err := uc.roomService.GetRoom(ctx, roomID)
	if err != nil {
		if errors.Is(err, entities.ErrRoomNotFound) {
			return nil, errors.Wrap(err, "room not found")
		}
		return nil, errors.Wrap(err, "failed to get room")
	}
	return room, nil
}
