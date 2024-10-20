package deleteroom

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

func (uc *UseCase) Execute(ctx context.Context, roomID, ownerID uuid.UUID) error {
	err := uc.roomService.DeleteRoom(ctx, roomID, ownerID)
	if err != nil {
		if errors.Is(err, entities.ErrRoomDeleteForbidden) {
			return errors.Wrap(err, "forbidden: cannot delete room")
		}
		if errors.Is(err, entities.ErrRoomNotFound) {
			return errors.Wrap(err, "room not found")
		}
		return errors.Wrap(err, "failed to delete room")
	}
	return nil
}
