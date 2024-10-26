package getmessages

import (
	"context"
	"sync"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UseCase struct {
	chatService    ChatService
	websiteService WebsiteService
}

func New(deps Deps) *UseCase {
	return &UseCase{
		chatService:    deps.ChatService,
		websiteService: deps.WebsiteService,
	}
}
func (uc *UseCase) Execute(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var wg sync.WaitGroup
	var existsErr, messagesErr error
	var exists bool
	var messages []*entities.Message

	wg.Add(2)

	go func() {
		defer wg.Done()
		exists, existsErr = uc.websiteService.RoomExists(ctx, roomID)
	}()

	go func() {
		defer wg.Done()
		messages, messagesErr = uc.chatService.GetMessages(ctx, roomID, limit, offset)
	}()

	wg.Wait()

	if existsErr != nil {
		return nil, errors.Wrap(existsErr, "failed to check room existence")
	}
	if !exists {
		return nil, entities.ErrRoomNotFound
	}

	if messagesErr != nil {
		return nil, errors.Wrap(messagesErr, "failed to get messages")
	}

	return messages, nil
}
