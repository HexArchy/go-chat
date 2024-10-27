package getmessages

import (
	"context"

	"github.com/pkg/errors"
)

// UseCase implements the get messages use case.
type UseCase struct {
	chatService ChatService
}

// New creates a new instance of the get messages use case.
func New(deps Deps) *UseCase {
	return &UseCase{
		chatService: deps.ChatService,
	}
}

// UseCase implements the get messages use case.
func (uc *UseCase) Execute(ctx context.Context, input MessagesInput) (*MessagesResponse, error) {
	// Validate limit.
	if input.Limit <= 0 {
		input.Limit = 50 // Default limit.
	} else if input.Limit > 100 {
		input.Limit = 100 // Max limit.
	}

	// Ensure offset is not negative.
	if input.Offset < 0 {
		input.Offset = 0
	}

	messages, err := uc.chatService.GetRoomMessages(ctx, input.RoomID, input.Limit, input.Offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get room messages")
	}

	response := &MessagesResponse{
		Messages: messages,
		Total:    len(messages),
	}

	return response, nil
}
