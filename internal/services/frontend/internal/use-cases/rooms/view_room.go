package rooms

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/chat"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ViewRoomUseCase defines the interface for viewing a room.
type ViewRoomUseCase interface {
	Execute(ctx context.Context, roomID string) (*entities.Room, []*entities.Message, error)
}

// viewRoomUseCase is the concrete implementation of ViewRoomUseCase.
type viewRoomUseCase struct {
	websiteClient *website.Client
	chatClient    *chat.Client
	logger        *zap.Logger
}

// NewViewRoomUseCase creates a new instance of ViewRoomUseCase.
func NewViewRoomUseCase(websiteClient *website.Client, chatClient *chat.Client, logger *zap.Logger) ViewRoomUseCase {
	return &viewRoomUseCase{
		websiteClient: websiteClient,
		chatClient:    chatClient,
		logger:        logger,
	}
}

// Execute retrieves room details and recent messages.
func (uc *viewRoomUseCase) Execute(ctx context.Context, roomIDStr string) (*entities.Room, []*entities.Message, error) {
	uc.logger.Debug("ViewRoomUseCase: viewing room",
		zap.String("room_id", roomIDStr))

	if roomIDStr == "" {
		return nil, nil, errors.New("room ID is required")
	}

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse room id")
	}

	room, err := uc.websiteClient.GetRoom(ctx, roomID)
	if err != nil {
		uc.logger.Error("ViewRoomUseCase: failed to get room", zap.Error(err))
		return nil, nil, errors.Wrap(err, "failed to get room")
	}

	messages, err := uc.chatClient.GetMessages(ctx, room.ID, 50, 0)
	if err != nil {
		uc.logger.Error("ViewRoomUseCase: failed to get messages", zap.Error(err))
		return nil, nil, errors.Wrap(err, "failed to get messages")
	}

	uc.logger.Debug("ViewRoomUseCase: retrieved room and messages successfully",
		zap.String("room_id", roomIDStr),
		zap.Int("messages_count", len(messages)))

	return room, messages, nil
}
