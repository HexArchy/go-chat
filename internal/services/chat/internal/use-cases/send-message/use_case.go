package sendmessage

import (
	"context"

	"github.com/pkg/errors"
)

// UseCase implements the send message use case.
type UseCase struct {
	chatService ChatService
}

// New creates a new instance of the send message use case.
func New(deps Deps) *UseCase {
	return &UseCase{
		chatService: deps.ChatService,
	}
}

// Execute sends a new message to a chat room.
func (uc *UseCase) Execute(ctx context.Context, input MessageInput) error {
	if input.Content == "" {
		return errors.New("message content cannot be empty")
	}

	if err := uc.chatService.HandleMessage(ctx, input.RoomID, input.UserID, input.Content, input.Event); err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	return nil
}
