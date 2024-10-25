package entities

import (
	"sync"

	"github.com/google/uuid"
)

type ChatConnection interface {
	SendEvent(event *ChatEvent) error
	Close() error
}

type RoomState struct {
	connections map[uuid.UUID]ChatConnection
	mu          sync.RWMutex
}

func NewRoomState() *RoomState {
	return &RoomState{
		connections: make(map[uuid.UUID]ChatConnection),
	}
}

func (r *RoomState) AddConnection(userID uuid.UUID, conn ChatConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections[userID] = conn
}

func (r *RoomState) RemoveConnection(userID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if conn, exists := r.connections[userID]; exists {
		conn.Close()
		delete(r.connections, userID)
	}
}

func (r *RoomState) BroadcastEvent(event *ChatEvent) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, conn := range r.connections {
		go func(c ChatConnection) {
			_ = c.SendEvent(event)
		}(conn)
	}
}

func (r *RoomState) GetConnections() map[uuid.UUID]ChatConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connections := make(map[uuid.UUID]ChatConnection, len(r.connections))
	for k, v := range r.connections {
		connections[k] = v
	}
	return connections
}
