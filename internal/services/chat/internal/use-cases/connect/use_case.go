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
	authService    AuthService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		websiteService: deps.WebsiteService,
		chatService:    deps.ChatService,
		authService:    deps.AuthService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, token string, roomID uuid.UUID, conn entities.ChatConnection) error {
	// Проверяем токен и получаем userID
	userID, err := uc.authService.ValidateToken(ctx, token)
	if err != nil {
		return errors.Wrap(err, "invalid token")
	}

	// Проверяем существование комнаты
	exists, err := uc.websiteService.RoomExists(ctx, roomID)
	if err != nil {
		return errors.Wrap(err, "failed to check room existence")
	}
	if !exists {
		return entities.ErrRoomNotFound
	}

	// Подключаем пользователя к чату
	if err := uc.chatService.Connect(ctx, roomID, userID, conn); err != nil {
		return errors.Wrap(err, "failed to connect to chat")
	}

	return nil
}
