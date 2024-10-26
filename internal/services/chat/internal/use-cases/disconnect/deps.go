package disconnect

import (
	"context"

	"github.com/google/uuid"
)

type ChatService interface {
	Disconnect(ctx context.Context, roomID, userID uuid.UUID) error
}

type Deps struct {
	ChatService ChatService
}
