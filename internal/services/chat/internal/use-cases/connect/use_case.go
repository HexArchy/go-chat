package connect

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UseCase struct {
	websiteService WebsiteService
	chatService    ChatService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		websiteService: deps.WebsiteService,
		chatService:    deps.ChatService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, userID, roomID uuid.UUID, conn entities.ChatConnection) error {
	exists, err := uc.websiteService.RoomExists(ctx, roomID)
	if err != nil {
		return errors.Wrap(err, "failed to check room existence")
	}
	if !exists {
		return entities.ErrRoomNotFound
	}

	if err := uc.chatService.Connect(ctx, roomID, userID, conn); err != nil {
		return errors.Wrap(err, "failed to connect to chat")
	}

	return nil
}
