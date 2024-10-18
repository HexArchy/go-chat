package storage

import (
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
)

type UserDTO struct {
	ID          string `gorm:"primaryKey"`
	Email       string `gorm:"uniqueIndex"`
	Password    string
	Name        string
	Nickname    string
	PhoneNumber string
	Age         int
	Bio         string
	Roles       []string `gorm:"type:text[]"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RefreshTokenDTO struct {
	ID        uint `gorm:"primaryKey"`
	UserID    string
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

func toUserDTO(user *entities.User) *UserDTO {
	return &UserDTO{
		ID:          user.ID,
		Email:       user.Email,
		Password:    user.Password,
		Name:        user.Name,
		Nickname:    user.Nickname,
		PhoneNumber: user.PhoneNumber,
		Age:         user.Age,
		Bio:         user.Bio,
		Roles:       rolesToStrings(user.Roles),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func fromUserDTO(dto *UserDTO) entities.User {
	return entities.User{
		ID:          dto.ID,
		Email:       dto.Email,
		Password:    dto.Password,
		Name:        dto.Name,
		Nickname:    dto.Nickname,
		PhoneNumber: dto.PhoneNumber,
		Age:         dto.Age,
		Bio:         dto.Bio,
		Roles:       stringsToRoles(dto.Roles),
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func rolesToStrings(roles []entities.Role) []string {
	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = string(role)
	}
	return result
}

func stringsToRoles(roles []string) []entities.Role {
	result := make([]entities.Role, len(roles))
	for i, role := range roles {
		result[i] = entities.Role(role)
	}
	return result
}
