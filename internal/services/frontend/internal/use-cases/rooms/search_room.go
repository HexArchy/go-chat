package rooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// SearchRoomsUseCase defines the interface for searching rooms.
type SearchRoomsUseCase interface {
	Execute(ctx context.Context, query string, limit, offset int32) ([]*entities.Room, error)
}

// searchRoomsUseCase is the concrete implementation of SearchRoomsUseCase.
type searchRoomsUseCase struct {
	websiteClient *website.Client
	logger        *zap.Logger
}

// NewSearchRoomsUseCase creates a new instance of SearchRoomsUseCase.
func NewSearchRoomsUseCase(websiteClient *website.Client, logger *zap.Logger) SearchRoomsUseCase {
	return &searchRoomsUseCase{
		websiteClient: websiteClient,
		logger:        logger,
	}
}

// Execute searches for rooms based on the query, limit, and offset.
func (uc *searchRoomsUseCase) Execute(ctx context.Context, query string, limit, offset int32) ([]*entities.Room, error) {
	uc.logger.Debug("SearchRoomsUseCase: searching rooms",
		zap.String("query", query),
		zap.Int32("limit", limit),
		zap.Int32("offset", offset))

	if len(query) > 0 && len(query) < 2 {
		return nil, errors.New("search query must be at least 2 characters long")
	}

	rooms, err := uc.websiteClient.SearchRooms(ctx, query, limit, offset)
	if err != nil {
		uc.logger.Error("SearchRoomsUseCase: failed to search rooms", zap.Error(err))
		return nil, errors.Wrap(err, "failed to search rooms")
	}

	uc.logger.Debug("SearchRoomsUseCase: search completed successfully",
		zap.Int("count", len(rooms)))

	return rooms, nil
}
