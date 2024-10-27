package storage

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/services/chat/internal/entities"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Storage defines methods for chat persistence operations.
type Storage struct {
	db *gorm.DB
}

// NewStorage creates a new instance of Storage.
func NewStorage(db *gorm.DB) *Storage {
	return &Storage{db: db}
}

// SaveMessage stores a new message in the database.
func (s *Storage) SaveMessage(ctx context.Context, msg *entities.Message) error {
	dto := &MessageDTO{
		ID:        msg.ID,
		RoomID:    msg.RoomID,
		UserID:    msg.UserID,
		Content:   msg.Content,
		CreatedAt: msg.Timestamp,
	}

	if err := s.db.WithContext(ctx).Create(dto).Error; err != nil {
		return errors.Wrap(err, "failed to save message")
	}

	return nil
}

// GetLastMessages retrieves messages from a specific room with pagination.
func (s *Storage) GetLastMessages(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*entities.Message, error) {
	var dtos []MessageDTO

	err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&dtos).
		Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to get last messages")
	}

	messages := make([]*entities.Message, len(dtos))
	for i, dto := range dtos {
		messages[i] = &entities.Message{
			ID:        dto.ID,
			RoomID:    dto.RoomID,
			UserID:    dto.UserID,
			Content:   dto.Content,
			Timestamp: dto.CreatedAt,
		}
	}

	return messages, nil
}

// AddParticipant adds a new participant to a room.
func (s *Storage) AddParticipant(ctx context.Context, roomID, userID uuid.UUID) error {
	isParticipant, err := s.IsParticipant(ctx, roomID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to check if participant exists")
	}
	if isParticipant {
		return nil
	}

	participant := &RoomParticipantDTO{
		RoomID:    roomID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	err = s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "room_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).
		Create(participant).
		Error

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return nil
			}
		}
		return errors.Wrap(err, "failed to add participant")
	}

	return nil
}

// RemoveParticipant removes a participant from a room.
func (s *Storage) RemoveParticipant(ctx context.Context, roomID, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Delete(&RoomParticipantDTO{})

	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to remove participant")
	}

	return nil
}

// IsParticipant checks if a user is a participant in a specific room.
func (s *Storage) IsParticipant(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&RoomParticipantDTO{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).
		Error

	if err != nil {
		return false, errors.Wrap(err, "failed to check participant status")
	}

	return count > 0, nil
}

// GetRoomParticipants retrieves all participants from a specific room.
func (s *Storage) GetRoomParticipants(ctx context.Context, roomID uuid.UUID) ([]uuid.UUID, error) {
	var participants []RoomParticipantDTO

	err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Find(&participants).
		Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to get room participants")
	}

	userIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		userIDs[i] = p.UserID
	}

	return userIDs, nil
}

// GetUserRooms retrieves all rooms where a user is a participant.
func (s *Storage) GetUserRooms(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var participants []RoomParticipantDTO

	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&participants).
		Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to get user rooms")
	}

	roomIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		roomIDs[i] = p.RoomID
	}

	return roomIDs, nil
}

// CleanupParticipants removes all participants from a specific room.
func (s *Storage) CleanupParticipants(ctx context.Context, roomID uuid.UUID) error {
	err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Delete(&RoomParticipantDTO{}).
		Error

	if err != nil {
		return errors.Wrap(err, "failed to cleanup room participants")
	}

	return nil
}
