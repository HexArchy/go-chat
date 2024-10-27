package getroomparticipants

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// UseCase implements the get room participants use case.
type UseCase struct {
	storage        Storage
	websiteService WebsiteService
}

// New creates a new instance of the get room participants use case.
func New(deps Deps) *UseCase {
	return &UseCase{
		storage:        deps.Storage,
		websiteService: deps.WebsiteService,
	}
}

// Execute retrieves all participants from a chat room.
func (uc *UseCase) Execute(ctx context.Context, roomID uuid.UUID) ([]uuid.UUID, error) {
	// Validate room existence.
	exists, err := uc.websiteService.RoomExists(ctx, roomID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify room existence")
	}
	if !exists {
		return nil, entities.ErrRoomNotFound
	}

	participants, err := uc.storage.GetRoomParticipants(ctx, roomID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get room participants")
	}

	return participants, nil
}
