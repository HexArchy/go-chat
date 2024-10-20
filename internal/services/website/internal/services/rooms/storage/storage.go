package storage

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Storage interface {
	CreateRoom(ctx context.Context, room *entities.Room) error
	GetRoomByID(ctx context.Context, roomID uuid.UUID) (*entities.Room, error)
	GetRoomsByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error)
	GetRoomsByName(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error)
	DeleteRoom(ctx context.Context, roomID uuid.UUID) error
}

type storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) Storage {
	return &storage{db: db}
}

func (s *storage) CreateRoom(ctx context.Context, room *entities.Room) error {
	dto := RoomToDTO(room)
	if err := s.db.WithContext(ctx).Create(&dto).Error; err != nil {
		return errors.Wrap(err, "failed to create room")
	}
	return nil
}

func (s *storage) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*entities.Room, error) {
	var dto Room
	if err := s.db.WithContext(ctx).First(&dto, "id = ?", roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrRoomNotFound
		}
		return nil, errors.Wrap(err, "failed to find room by ID")
	}
	return DTOToRoom(&dto), nil
}

func (s *storage) GetRoomsByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error) {
	var dtos []Room
	if err := s.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&dtos).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find rooms by owner ID")
	}
	return DTOsToRooms(dtos), nil
}

func (s *storage) GetRoomsByName(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error) {
	var dtos []Room
	if err := s.db.WithContext(ctx).
		Where("name LIKE ?", "%"+name+"%").
		Limit(limit).
		Offset(offset).
		Find(&dtos).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find rooms by name")
	}
	return DTOsToRooms(dtos), nil
}

func (s *storage) DeleteRoom(ctx context.Context, roomID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&Room{}, "id = ?", roomID).Error; err != nil {
		return errors.Wrap(err, "failed to delete room")
	}
	return nil
}
