package createroom

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

func (uc *UseCase) Execute(ctx context.Context, name string, ownerID uuid.UUID) (*entities.Room, error) {
	room, err := uc.roomService.CreateRoom(ctx, name, ownerID)
	if err != nil {
		if errors.Is(err, entities.ErrRoomAlreadyExists) {
			return nil, errors.Wrap(err, "room already exists")
		}
		return nil, errors.Wrap(err, "failed to create room")
	}
	return room, nil
}
