package disconnect

import (
	"context"

	"github.com/pkg/errors"
)

// UseCase implements the disconnect use case.
type UseCase struct {
	chatService ChatService
}

// New creates a new instance of the disconnect use case.
func New(deps Deps) *UseCase {
	return &UseCase{
		chatService: deps.ChatService,
	}
}

// Execute performs the disconnection of a user from a chat room.
func (uc *UseCase) Execute(ctx context.Context, input DisconnectInput) error {
	if err := uc.chatService.Disconnect(ctx, input.RoomID, input.UserID); err != nil {
		return errors.Wrap(err, "failed to disconnect from chat room")
	}
	return nil
}
