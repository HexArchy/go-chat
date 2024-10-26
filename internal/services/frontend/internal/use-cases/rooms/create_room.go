package rooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// CreateRoomUseCase defines the interface for creating a new room.
type CreateRoomUseCase interface {
	Execute(ctx context.Context, name, ownerID string) (*entities.Room, error)
}

// createRoomUseCase is the concrete implementation of CreateRoomUseCase.
type createRoomUseCase struct {
	websiteClient *website.Client
	logger        *zap.Logger
}

// NewCreateRoomUseCase creates a new instance of CreateRoomUseCase.
func NewCreateRoomUseCase(websiteClient *website.Client, logger *zap.Logger) CreateRoomUseCase {
	return &createRoomUseCase{
		websiteClient: websiteClient,
		logger:        logger,
	}
}

// Execute creates a new room with the specified name and owner.
func (uc *createRoomUseCase) Execute(ctx context.Context, name, ownerIDStr string) (*entities.Room, error) {
	uc.logger.Debug("CreateRoomUseCase: creating new room",
		zap.String("name", name),
		zap.String("owner_id", ownerIDStr))

	if name == "" {
		return nil, errors.New("room name is required")
	}

	if ownerIDStr == "" {
		return nil, errors.New("owner ID is required")
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse owner id")
	}

	room, err := uc.websiteClient.CreateRoom(ctx, name, ownerID)
	if err != nil {
		uc.logger.Error("CreateRoomUseCase: failed to create room", zap.Error(err))
		return nil, errors.Wrap(err, "failed to create room")
	}

	uc.logger.Debug("CreateRoomUseCase: room created successfully",
		zap.String("room_id", room.ID.String()))

	return room, nil
}
