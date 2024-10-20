package getownerrooms

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

func (uc *UseCase) Execute(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error) {

	rooms, err := uc.roomService.GetOwnerRooms(ctx, ownerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get owner's rooms")
	}
	return rooms, nil
}
