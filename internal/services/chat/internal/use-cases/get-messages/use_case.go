package getmessages

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UseCase struct {
	chatService    ChatService
	websiteService WebsiteService
	authService    AuthService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		chatService:    deps.ChatService,
		websiteService: deps.WebsiteService,
		authService:    deps.AuthService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, token string, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	_, err := uc.authService.ValidateToken(ctx, token)
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	exists, err := uc.websiteService.RoomExists(ctx, roomID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check room existence")
	}
	if !exists {
		return nil, entities.ErrRoomNotFound
	}

	messages, err := uc.chatService.GetMessages(ctx, roomID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get messages")
	}

	return messages, nil
}
