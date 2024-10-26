package rooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ListRoomsUseCase defines the interface for listing user-owned rooms.
type ListOwnRoomsUseCase interface {
	Execute(ctx context.Context, ownerID string) ([]*entities.Room, error)
}

// listRoomsUseCase is the concrete implementation of ListRoomsUseCase.
type listOwnRoomsUseCase struct {
	websiteClient *website.Client
	logger        *zap.Logger
}

// NewListRoomsUseCase creates a new instance of ListRoomsUseCase.
func NewOwnListRoomsUseCase(websiteClient *website.Client, logger *zap.Logger) ListOwnRoomsUseCase {
	return &listOwnRoomsUseCase{
		websiteClient: websiteClient,
		logger:        logger,
	}
}

// Execute retrieves the list of rooms owned by the user.
func (uc *listOwnRoomsUseCase) Execute(ctx context.Context, ownerIDStr string) ([]*entities.Room, error) {
	uc.logger.Debug("ListRoomsUseCase: fetching owner rooms",
		zap.String("owner_id", ownerIDStr))

	// Validate ownerID.
	if ownerIDStr == "" {
		return nil, errors.New("owner ID is required")
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		return nil, errors.Wrap(err, "parse owner id")
	}

	rooms, err := uc.websiteClient.GetOwnerRooms(ctx, ownerID)
	if err != nil {
		uc.logger.Error("ListRoomsUseCase: failed to get owner rooms", zap.Error(err))
		return nil, errors.Wrap(err, "failed to get owner rooms")
	}

	uc.logger.Debug("ListRoomsUseCase: fetched rooms successfully",
		zap.Int("count", len(rooms)))

	return rooms, nil
}
