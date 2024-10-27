package storage

import (
	"time"

	"github.com/google/uuid"
)

// MessageDTO represents a chat message in the database.
type MessageDTO struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	RoomID    uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Content   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (MessageDTO) TableName() string {
	return "chat_messages"
}

// RoomParticipantDTO represents a room participant in the database.
type RoomParticipantDTO struct {
	RoomID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_room_user"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_room_user"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (RoomParticipantDTO) TableName() string {
	return "chat_room_participants"
}
