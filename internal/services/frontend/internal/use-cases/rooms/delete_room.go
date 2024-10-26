package rooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// DeleteRoomUseCase defines the interface for deleting a room.
type DeleteRoomUseCase interface {
	Execute(ctx context.Context, roomID, ownerID string) error
}

// deleteRoomUseCase is the concrete implementation of DeleteRoomUseCase.
type deleteRoomUseCase struct {
	websiteClient *website.Client
	logger        *zap.Logger
}

// NewDeleteRoomUseCase creates a new instance of DeleteRoomUseCase.
func NewDeleteRoomUseCase(websiteClient *website.Client, logger *zap.Logger) DeleteRoomUseCase {
	return &deleteRoomUseCase{
		websiteClient: websiteClient,
		logger:        logger,
	}
}

// Execute deletes the specified room if the requester is the owner.
func (uc *deleteRoomUseCase) Execute(ctx context.Context, roomIDStr, ownerIDStr string) error {
	uc.logger.Debug("DeleteRoomUseCase: deleting room",
		zap.String("room_id", roomIDStr),
		zap.String("owner_id", ownerIDStr))

	if roomIDStr == "" || ownerIDStr == "" {
		return errors.New("room ID and owner ID are required")
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return errors.Wrap(err, "parse room id")
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		return errors.Wrap(err, "parse owner id")
	}

	// Call the website client to delete the room.
	err = uc.websiteClient.DeleteRoom(ctx, roomID, ownerID)
	if err != nil {
		uc.logger.Error("DeleteRoomUseCase: failed to delete room", zap.Error(err))
		return errors.Wrap(err, "failed to delete room")
	}

	uc.logger.Debug("DeleteRoomUseCase: room deleted successfully",
		zap.String("room_id", roomIDStr))

	return nil
}
