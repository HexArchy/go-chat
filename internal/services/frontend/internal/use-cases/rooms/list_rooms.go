package rooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ListRoomsUseCase defines the interface for listing all rooms with pagination.
type ListRoomsUseCase interface {
	Execute(ctx context.Context, limit, offset int32) ([]*entities.Room, error)
}

// listRoomsUseCase is the concrete implementation of ListRoomsUseCase.
type listRoomsUseCase struct {
	websiteClient *website.Client
	logger        *zap.Logger
}

// NewListRoomsUseCase creates a new instance of ListRoomsUseCase.
func NewListRoomsUseCase(websiteClient *website.Client, logger *zap.Logger) ListRoomsUseCase {
	return &listRoomsUseCase{
		websiteClient: websiteClient,
		logger:        logger,
	}
}

// Execute retrieves a paginated list of all rooms.
func (uc *listRoomsUseCase) Execute(ctx context.Context, limit, offset int32) ([]*entities.Room, error) {
	uc.logger.Debug("ListRoomsUseCase: fetching all rooms",
		zap.Int32("limit", limit), zap.Int32("offset", offset))

	// Ensure limit and offset are valid
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}

	rooms, err := uc.websiteClient.GetAllRooms(ctx, limit, offset)
	if err != nil {
		uc.logger.Error("ListRoomsUseCase: failed to get all rooms", zap.Error(err))
		return nil, errors.Wrap(err, "failed to get all rooms")
	}

	uc.logger.Debug("ListRoomsUseCase: fetched rooms successfully",
		zap.Int("count", len(rooms)))

	return rooms, nil
}
