package disconnect

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UseCase struct {
	chatService ChatService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		chatService: deps.ChatService,
	}
}

func (uc *UseCase) Execute(ctx context.Context, userID, roomID uuid.UUID) error {
	if err := uc.chatService.Disconnect(ctx, roomID, userID); err != nil {
		return errors.Wrap(err, "failed to disconnect from chat")
	}

	return nil
}
