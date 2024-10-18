package storage

import (
	"context"

	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateUser(ctx context.Context, user *entities.User) error {
	userDTO := toUserDTO(user)
	result := s.db.WithContext(ctx).Create(userDTO)
	if result.Error != nil {
		return result.Error
	}
	*user = fromUserDTO(userDTO)
	return nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var userDTO UserDTO
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&userDTO)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	user := fromUserDTO(&userDTO)
	return &user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	var userDTO UserDTO
	result := s.db.WithContext(ctx).First(&userDTO, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	user := fromUserDTO(&userDTO)
	return &user, nil
}

func (s *Storage) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&RefreshTokenDTO{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to delete refresh tokens")
	}
	return nil
}

func (s *Storage) UpdateUser(ctx context.Context, user *entities.User) error {
	userDTO := toUserDTO(user)
	result := s.db.WithContext(ctx).Save(userDTO)
	if result.Error != nil {
		return result.Error
	}
	*user = fromUserDTO(userDTO)
	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&UserDTO{}, "id = ?", id)
	return result.Error
}

func (s *Storage) StoreRefreshToken(ctx context.Context, userID, token string, ttl time.Duration) error {
	refreshToken := RefreshTokenDTO{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(ttl),
	}
	result := s.db.WithContext(ctx).Create(&refreshToken)
	return result.Error
}

func (s *Storage) ValidateRefreshToken(ctx context.Context, userID, token string) (bool, error) {
	var refreshToken RefreshTokenDTO
	result := s.db.WithContext(ctx).Where("user_id = ? AND token = ? AND expires_at > ?", userID, token, time.Now()).First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, result.Error
	}
	return true, nil
}

func (s *Storage) RevokeRefreshToken(ctx context.Context, userID, token string) error {
	result := s.db.WithContext(ctx).Where("user_id = ? AND token = ?", userID, token).Delete(&RefreshTokenDTO{})
	return result.Error
}
