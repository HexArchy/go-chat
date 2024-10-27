// internal/services/chat/internal/services/chat/chat.go
package chat

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	storage     Storage
	logger      *zap.Logger
	rooms       map[uuid.UUID]*entities.Room
	mu          sync.RWMutex
	cleanupTick *time.Ticker
}

func NewService(deps Deps, logger *zap.Logger) *Service {
	s := &Service{
		storage: deps.Storage,
		logger:  logger,
		rooms:   make(map[uuid.UUID]*entities.Room),
	}

	s.startCleanupTicker()
	return s
}

func (s *Service) Connect(ctx context.Context, roomID, userID uuid.UUID, conn entities.Connection) error {
	room := s.getOrCreateRoom(roomID)

	if err := s.storage.AddParticipant(ctx, roomID, userID); err != nil {
		s.logger.Error("Failed to add participant",
			zap.Error(err),
			zap.String("room_id", roomID.String()),
			zap.String("user_id", userID.String()),
		)
		return errors.Wrap(err, "failed to add participant")
	}

	room.AddConnection(userID, conn)

	userConnectEvent := &entities.Event{
		Type:      entities.EventUserConnected,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	room.BroadcastEvent(userConnectEvent, &userID)

	s.logger.Info("User connected to room",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()),
		zap.String("connection_type", string(conn.Type())),
	)

	return nil
}

func (s *Service) Disconnect(ctx context.Context, roomID, userID uuid.UUID) error {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	disconnectEvent := &entities.Event{
		Type:      entities.EventUserDisconnected,
		RoomID:    roomID,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	room.BroadcastEvent(disconnectEvent, nil)

	if err := s.storage.RemoveParticipant(ctx, roomID, userID); err != nil {
		s.logger.Error("Failed to remove participant",
			zap.Error(err),
			zap.String("room_id", roomID.String()),
			zap.String("user_id", userID.String()),
		)
	}

	room.RemoveConnection(userID)

	s.cleanupRoomIfEmpty(roomID)

	s.logger.Info("User disconnected from room",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}

func (s *Service) HandleMessage(ctx context.Context, roomID, userID uuid.UUID, content string, event *entities.Event) error {
	isParticipant, err := s.storage.IsParticipant(ctx, roomID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to verify participant")
	}
	if !isParticipant {
		return errors.New("user is not a participant of this room")
	}

	room := s.getRoom(roomID)
	if room == nil {
		return entities.ErrRoomNotFound
	}

	msg := &entities.Message{
		ID:        uuid.New(),
		RoomID:    roomID,
		UserID:    userID,
		Content:   content,
		Timestamp: time.Now(),
	}

	if err := s.storage.SaveMessage(ctx, msg); err != nil {
		return errors.Wrap(err, "failed to save message")
	}

	if event == nil {
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			return errors.Wrap(err, "failed to marshal message")
		}

		event = &entities.Event{
			Type:      entities.EventNewMessage,
			RoomID:    roomID,
			UserID:    userID,
			Payload:   msgJSON,
			Timestamp: time.Now(),
		}
	}

	room.BroadcastEvent(event, nil)

	s.logger.Debug("Message handled",
		zap.String("room_id", roomID.String()),
		zap.String("user_id", userID.String()),
		zap.String("message_id", msg.ID.String()),
	)

	return nil
}

func (s *Service) GetRoomMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	messages, err := s.storage.GetLastMessages(ctx, roomID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get messages")
	}

	room := s.getRoom(roomID)
	if room == nil {
		return nil, entities.ErrRoomNotFound
	}

	return messages, nil
}

// Internal helper methods.
func (s *Service) getOrCreateRoom(roomID uuid.UUID) *entities.Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.rooms[roomID]; exists {
		return room
	}

	room := entities.NewRoom(roomID, s.logger)
	s.rooms[roomID] = room
	return room
}

func (s *Service) getRoom(roomID uuid.UUID) *entities.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rooms[roomID]
}

func (s *Service) cleanupRoomIfEmpty(roomID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.rooms[roomID]; exists {
		if room.IsEmpty() {
			room.CleanupConnections()
			delete(s.rooms, roomID)
			s.logger.Info("Room cleaned up due to no active connections",
				zap.String("room_id", roomID.String()),
			)
		}
	}
}

func (s *Service) startCleanupTicker() {
	s.cleanupTick = time.NewTicker(5 * time.Minute)
	go func() {
		for range s.cleanupTick.C {
			s.cleanupInactiveRooms()
		}
	}()
}

func (s *Service) cleanupInactiveRooms() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for roomID, room := range s.rooms {
		if room.IsEmpty() {
			room.CleanupConnections()
			delete(s.rooms, roomID)
			s.logger.Info("Room cleaned up during periodic cleanup",
				zap.String("room_id", roomID.String()),
			)
		}
	}
}

// Cleanup performs cleanup of all rooms and connections
func (s *Service) Cleanup() {
	if s.cleanupTick != nil {
		s.cleanupTick.Stop()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for roomID, room := range s.rooms {
		room.CleanupConnections()
		delete(s.rooms, roomID)
	}

	s.logger.Info("Chat service cleanup completed")
}
