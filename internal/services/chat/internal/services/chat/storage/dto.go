package storage

import (
	"time"

	"github.com/google/uuid"
)

type MessageDTO struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	RoomID    uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (MessageDTO) TableName() string {
	return "chat_messages"
}

type ParticipantDTO struct {
	RoomID   uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID   uuid.UUID `gorm:"type:uuid;primary_key"`
	JoinedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (ParticipantDTO) TableName() string {
	return "chat_participants"
}
