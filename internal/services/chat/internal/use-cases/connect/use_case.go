package connect

import (
	"context"

	"github.com/pkg/errors"
)

// UseCase implements the connect use case.
type UseCase struct {
	websiteService WebsiteService
	chatService    ChatService
}

// New creates a new instance of the connect use case.
func New(deps Deps) *UseCase {
	return &UseCase{
		websiteService: deps.WebsiteService,
		chatService:    deps.ChatService,
	}
}

// Execute performs the connection of a user to a chat room.
func (uc *UseCase) Execute(ctx context.Context, input ConnectInput) error {
	// Connect to chat room
	if err := uc.chatService.Connect(ctx, input.RoomID, input.UserID, input.Connection); err != nil {
		return errors.Wrap(err, "failed to connect to chat room")
	}

	return nil
}
