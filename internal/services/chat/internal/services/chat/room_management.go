package chat

import (
	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
)

func (s *Service) getOrCreateRoom(roomID uuid.UUID) *entities.RoomState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.rooms[roomID]; exists {
		return room
	}

	room := entities.NewRoomState()
	s.rooms[roomID] = room
	return room
}

func (s *Service) getRoom(roomID uuid.UUID) *entities.RoomState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rooms[roomID]
}

func (s *Service) removeRoom(roomID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.rooms[roomID]; exists {
		for userID, conn := range room.GetConnections() {
			conn.Close()
			room.RemoveConnection(userID)
		}
		delete(s.rooms, roomID)
	}
}

func (s *Service) CleanupRooms() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for roomID, room := range s.rooms {
		for userID, conn := range room.GetConnections() {
			conn.Close()
			room.RemoveConnection(userID)
		}
		delete(s.rooms, roomID)
	}
}
