package entities

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Room struct {
	ID           uuid.UUID
	connections  sync.Map // map[uuid.UUID]Connection.
	logger       *zap.Logger
	lastActivity time.Time
	mu           sync.RWMutex
}

func NewRoom(id uuid.UUID, logger *zap.Logger) *Room {
	return &Room{
		ID:           id,
		logger:       logger,
		lastActivity: time.Now(),
	}
}

func (r *Room) AddConnection(userID uuid.UUID, conn Connection) {
	if existingConn, loaded := r.connections.LoadAndDelete(userID); loaded {
		if existing, ok := existingConn.(Connection); ok {
			r.logger.Debug("Closing existing connection",
				zap.String("room_id", r.ID.String()),
				zap.String("user_id", userID.String()),
			)
			existing.Close()
		}
	}

	r.logger.Debug("Adding new connection",
		zap.String("room_id", r.ID.String()),
		zap.String("user_id", userID.String()),
	)
	r.connections.Store(userID, conn)
	r.updateLastActivity()
}

func (r *Room) RemoveConnection(userID uuid.UUID) {
	if conn, loaded := r.connections.LoadAndDelete(userID); loaded {
		if connection, ok := conn.(Connection); ok {
			r.logger.Debug("Removing connection",
				zap.String("room_id", r.ID.String()),
				zap.String("user_id", userID.String()),
			)
			connection.Close()
		}
	}
	r.updateLastActivity()
}

func (r *Room) BroadcastEvent(event *Event, excludeUserID *uuid.UUID) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		r.logger.Error("Failed to marshal event",
			zap.Error(err),
			zap.String("room_id", r.ID.String()),
			zap.String("event_type", string(event.Type)),
		)
		return
	}

	r.connections.Range(func(key, value interface{}) bool {
		userID := key.(uuid.UUID)
		if excludeUserID != nil && userID == *excludeUserID {
			return true
		}

		conn := value.(Connection)
		go func(c Connection, uid uuid.UUID) {
			if err := c.Send(eventJSON); err != nil {
				if err == ErrConnectionClosed {
					r.logger.Debug("Removing closed connection during broadcast",
						zap.String("room_id", r.ID.String()),
						zap.String("user_id", uid.String()),
					)
					r.RemoveConnection(uid)
				} else {
					r.logger.Error("Failed to send event",
						zap.Error(err),
						zap.String("room_id", r.ID.String()),
						zap.String("user_id", uid.String()),
						zap.String("event_type", string(event.Type)),
					)
				}
			}
		}(conn, userID)
		return true
	})
	r.updateLastActivity()
}

func (r *Room) IsEmpty() bool {
	empty := true
	r.connections.Range(func(key, value interface{}) bool {
		conn := value.(Connection)
		if !conn.IsClosed() {
			empty = false
			return false
		}
		return true
	})
	return empty
}

func (r *Room) CleanupConnections() {
	r.connections.Range(func(key, value interface{}) bool {
		userID := key.(uuid.UUID)
		if conn, ok := value.(Connection); ok {
			r.logger.Debug("Cleaning up connection",
				zap.String("room_id", r.ID.String()),
				zap.String("user_id", userID.String()),
			)
			conn.Close()
		}
		r.connections.Delete(key)
		return true
	})
}

// GetConnectionCount возвращает количество активных соединений
func (r *Room) GetConnectionCount() int {
	count := 0
	r.connections.Range(func(key, value interface{}) bool {
		conn := value.(Connection)
		if !conn.IsClosed() {
			count++
		}
		return true
	})
	return count
}

// CheckConnection проверяет существование и валидность соединения
func (r *Room) CheckConnection(userID uuid.UUID) bool {
	if conn, ok := r.connections.Load(userID); ok {
		if connection, ok := conn.(Connection); ok {
			return !connection.IsClosed()
		}
	}
	return false
}

func (r *Room) GetLastActivity() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastActivity
}

func (r *Room) updateLastActivity() {
	r.mu.Lock()
	r.lastActivity = time.Now()
	r.mu.Unlock()
}

func (r *Room) GetParticipants() []uuid.UUID {
	var participants []uuid.UUID
	r.connections.Range(func(key, value interface{}) bool {
		userID := key.(uuid.UUID)
		conn := value.(Connection)
		if !conn.IsClosed() {
			participants = append(participants, userID)
		}
		return true
	})
	return participants
}
