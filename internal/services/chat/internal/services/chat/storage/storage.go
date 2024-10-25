package storage

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Storage interface {
	// Messages.
	CreateMessage(ctx context.Context, message *entities.Message) error
	GetMessage(ctx context.Context, messageID uuid.UUID) (*entities.Message, error)
	GetRoomMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error)

	// Participants.
	AddParticipant(ctx context.Context, roomID, userID uuid.UUID) error
	RemoveParticipant(ctx context.Context, roomID, userID uuid.UUID) error
	GetRoomParticipants(ctx context.Context, roomID uuid.UUID) ([]*entities.ChatParticipant, error)
	IsParticipant(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
}

type storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) Storage {
	return &storage{db: db}
}

func (s *storage) CreateMessage(ctx context.Context, message *entities.Message) error {
	dto := &MessageDTO{
		ID:        message.ID,
		RoomID:    message.RoomID,
		UserID:    message.UserID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}

	if err := s.db.WithContext(ctx).Create(dto).Error; err != nil {
		return errors.Wrap(err, "failed to create message")
	}

	return nil
}

func (s *storage) GetMessage(ctx context.Context, messageID uuid.UUID) (*entities.Message, error) {
	var dto MessageDTO
	if err := s.db.WithContext(ctx).First(&dto, "id = ?", messageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrMessageNotFound
		}
		return nil, errors.Wrap(err, "failed to get message")
	}

	return &entities.Message{
		ID:        dto.ID,
		RoomID:    dto.RoomID,
		UserID:    dto.UserID,
		Content:   dto.Content,
		CreatedAt: dto.CreatedAt,
	}, nil
}

func (s *storage) GetRoomMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	var dtos []MessageDTO
	if err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dtos).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get room messages")
	}

	messages := make([]*entities.Message, len(dtos))
	for i, dto := range dtos {
		messages[i] = &entities.Message{
			ID:        dto.ID,
			RoomID:    dto.RoomID,
			UserID:    dto.UserID,
			Content:   dto.Content,
			CreatedAt: dto.CreatedAt,
		}
	}

	return messages, nil
}

func (s *storage) AddParticipant(ctx context.Context, roomID, userID uuid.UUID) error {
	participant := &ParticipantDTO{
		RoomID:   roomID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(participant).Error; err != nil {
		return errors.Wrap(err, "failed to add participant")
	}

	return nil
}

func (s *storage) RemoveParticipant(ctx context.Context, roomID, userID uuid.UUID) error {
	if err := s.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&ParticipantDTO{}).Error; err != nil {
		return errors.Wrap(err, "failed to remove participant")
	}

	return nil
}

func (s *storage) GetRoomParticipants(ctx context.Context, roomID uuid.UUID) ([]*entities.ChatParticipant, error) {
	var dtos []ParticipantDTO
	if err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Find(&dtos).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get room participants")
	}

	participants := make([]*entities.ChatParticipant, len(dtos))
	for i, dto := range dtos {
		participants[i] = &entities.ChatParticipant{
			RoomID:   dto.RoomID,
			UserID:   dto.UserID,
			JoinedAt: dto.JoinedAt,
		}
	}

	return participants, nil
}

func (s *storage) IsParticipant(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&ParticipantDTO{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error; err != nil {
		return false, errors.Wrap(err, "failed to check participant")
	}

	return count > 0, nil
}
