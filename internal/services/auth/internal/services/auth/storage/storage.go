package storage

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Storage interface {
	CreateToken(ctx context.Context, token *entities.Token) error
	GetToken(ctx context.Context, tokenString string) (*entities.Token, error)
	DeleteToken(ctx context.Context, tokenString string) error
	DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error
}

type storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) Storage {
	return &storage{
		db: db,
	}
}

func (s *storage) CreateToken(ctx context.Context, token *entities.Token) error {
	dto := Token{
		Token:     token.Token,
		UserID:    token.UserID,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
		UpdatedAt: token.UpdatedAt,
	}
	if err := s.db.WithContext(ctx).Create(&dto).Error; err != nil {
		return errors.Wrap(err, "failed to create token")
	}
	return nil
}

func (s *storage) GetToken(ctx context.Context, tokenString string) (*entities.Token, error) {
	var dto Token
	if err := s.db.WithContext(ctx).
		Preload("User").
		First(&dto, "token = ?", tokenString).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrTokenNotFound
		}
		return nil, errors.Wrap(err, "failed to get token")
	}
	return dtoToEntity(&dto), nil
}

func (s *storage) DeleteToken(ctx context.Context, tokenString string) error {
	if err := s.db.WithContext(ctx).Where("token = ?", tokenString).Delete(&Token{}).Error; err != nil {
		return errors.Wrap(err, "failed to delete token")
	}
	return nil
}

func (s *storage) DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&Token{}).Error; err != nil {
		return errors.Wrap(err, "failed to delete tokens for user")
	}
	return nil
}
