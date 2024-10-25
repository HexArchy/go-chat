package disconnect

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UseCase struct {
	chatService ChatService
	authService AuthService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		chatService: deps.ChatService,
		authService: deps.AuthService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, token string, roomID uuid.UUID) error {
	userID, err := uc.authService.ValidateToken(ctx, token)
	if err != nil {
		return errors.Wrap(err, "invalid token")
	}

	if err := uc.chatService.Disconnect(ctx, roomID, userID); err != nil {
		return errors.Wrap(err, "failed to disconnect from chat")
	}

	return nil
}
