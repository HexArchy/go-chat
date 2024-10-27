package getmessages

import (
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

// MessagesInput represents the input data for getting messages.
type MessagesInput struct {
	RoomID uuid.UUID
	Limit  int
	Offset int
}

// MessagesResponse represents the response structure for messages.
type MessagesResponse struct {
	Messages []*entities.Message `json:"messages"`
	Total    int                 `json:"total"`
}
