package storage

import (
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	userstorage "github.com/HexArch/go-chat/internal/services/auth/internal/services/user/storage"
	"github.com/google/uuid"
)

type Token struct {
	Token     string           `gorm:"primaryKey;column:token"`
	UserID    uuid.UUID        `gorm:"column:user_id;type:uuid"`
	User      userstorage.User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ExpiresAt time.Time        `gorm:"column:expires_at"`
	CreatedAt time.Time        `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime;column:updated_at"`
}

func dtoToEntity(dto *Token) *entities.Token {
	return &entities.Token{
		Token:     dto.Token,
		UserID:    dto.UserID,
		ExpiresAt: dto.ExpiresAt,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}
