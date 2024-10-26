package sendmessage

import (
	"context"
	"strings"

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

func (uc *UseCase) Execute(ctx context.Context, userID, roomID uuid.UUID, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("message content cannot be empty")
	}

	if err := uc.chatService.SendMessage(ctx, roomID, userID, content); err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}
