package storage

import (
	"time"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;unique"`
	OwnerID   uuid.UUID `gorm:"column:owner_id;type:uuid"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func RoomToDTO(room *entities.Room) *Room {
	return &Room{
		ID:        room.ID,
		Name:      room.Name,
		OwnerID:   room.OwnerID,
		CreatedAt: room.CreatedAt,
		UpdatedAt: room.UpdatedAt,
	}
}

func DTOToRoom(dto *Room) *entities.Room {
	return &entities.Room{
		ID:        dto.ID,
		Name:      dto.Name,
		OwnerID:   dto.OwnerID,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

func DTOsToRooms(dtos []Room) []*entities.Room {
	rooms := make([]*entities.Room, len(dtos))
	for i, dto := range dtos {
		rooms[i] = DTOToRoom(&dto)
	}
	return rooms
}
