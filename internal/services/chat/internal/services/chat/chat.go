package chat

import (
	"context"
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	storage MessageStorage
	website WebsiteClient
	logger  *zap.Logger

	rooms map[uuid.UUID]*entities.RoomState
	mu    sync.RWMutex
}

func New(deps Deps, logger *zap.Logger) *Service {
	return &Service{
		storage: deps.MessageStorage,
		website: deps.WebsiteClient,
		logger:  logger,
		rooms:   make(map[uuid.UUID]*entities.RoomState),
	}
}

func (s *Service) Connect(ctx context.Context, roomID, userID uuid.UUID, conn entities.ChatConnection) error {
	exists, err := s.website.RoomExists(ctx, roomID)
	if err != nil {
		return errors.Wrap(err, "failed to check room existence")
	}
	if !exists {
		return entities.ErrRoomNotFound
	}

	room := s.getOrCreateRoom(roomID)

	room.AddConnection(userID, conn)

	if err := s.storage.AddParticipant(ctx, roomID, userID); err != nil {
		s.logger.Error("Failed to add participant to storage",
			zap.Error(err),
			zap.String("room_id", roomID.String()),
			zap.String("user_id", userID.String()),
		)
		return errors.Wrap(err, "failed to add participant")
	}

	event := &entities.ChatEvent{
		RoomID:    roomID,
		UserID:    userID,
		Type:      entities.EventTypeUserJoined,
		Timestamp: time.Now(),
	}
	room.BroadcastEvent(event)

	s.logger.Info("User connected to chat room",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}

func (s *Service) Disconnect(ctx context.Context, roomID, userID uuid.UUID) error {
	room := s.getRoom(roomID)
	if room == nil {
		return nil
	}

	room.RemoveConnection(userID)

	if err := s.storage.RemoveParticipant(ctx, roomID, userID); err != nil {
		s.logger.Error("Failed to remove participant from storage",
			zap.Error(err),
			zap.String("room_id", roomID.String()),
			zap.String("user_id", userID.String()),
		)
		return errors.Wrap(err, "failed to remove participant")
	}

	event := &entities.ChatEvent{
		RoomID:    roomID,
		UserID:    userID,
		Type:      entities.EventTypeUserLeft,
		Timestamp: time.Now(),
	}
	room.BroadcastEvent(event)

	s.logger.Info("User disconnected from chat room",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()),
	)

	if len(room.GetConnections()) == 0 {
		s.removeRoom(roomID)
	}

	return nil
}

func (s *Service) SendMessage(ctx context.Context, roomID, userID uuid.UUID, content string) error {
	isParticipant, err := s.storage.IsParticipant(ctx, roomID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to check participant")
	}
	if !isParticipant {
		return errors.New("user is not a participant of this chat")
	}

	message := &entities.Message{
		ID:        uuid.New(),
		RoomID:    roomID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.storage.CreateMessage(ctx, message); err != nil {
		s.logger.Error("Failed to create message in storage",
			zap.Error(err),
			zap.String("room_id", roomID.String()),
			zap.String("user_id", userID.String()),
		)
		return errors.Wrap(err, "failed to create message")
	}

	room := s.getRoom(roomID)
	if room == nil {
		s.logger.Warn("Room not found in memory",
			zap.String("room_id", roomID.String()),
		)
		return nil
	}

	event := &entities.ChatEvent{
		RoomID:    roomID,
		UserID:    userID,
		Type:      entities.EventTypeNewMessage,
		Message:   message,
		Timestamp: time.Now(),
	}
	room.BroadcastEvent(event)

	s.logger.Debug("Message sent to chat room",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()),
		zap.String("message_id", message.ID.String()),
	)

	return nil
}

func (s *Service) GetMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	messages, err := s.storage.GetRoomMessages(ctx, roomID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get room messages",
			zap.Error(err),
			zap.String("room_id", roomID.String()),
			zap.Int("limit", limit),
			zap.Int("offset", offset),
		)
		return nil, errors.Wrap(err, "failed to get room messages")
	}

	return messages, nil
}
