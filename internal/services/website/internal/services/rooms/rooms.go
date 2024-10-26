package rooms

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/services/website/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Service interface {
	CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*entities.Room, error)
	GetRoom(ctx context.Context, roomID uuid.UUID) (*entities.Room, error)
	GetOwnerRooms(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error)
	SearchRooms(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error)
	DeleteRoom(ctx context.Context, roomID, ownerID uuid.UUID) error
	GetAllRooms(ctx context.Context, limit, offset int) ([]*entities.Room, error)
}

type service struct {
	roomStorage RoomStorage
}

func NewService(deps Deps) Service {
	return &service{
		roomStorage: deps.RoomStorage,
	}
}

func (s *service) CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*entities.Room, error) {
	// Ensure room name is unique.
	existingRooms, err := s.roomStorage.GetRoomsByName(ctx, name, 1, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check for existing room")
	}
	if len(existingRooms) > 0 {
		return nil, entities.ErrRoomAlreadyExists
	}

	// Create room.
	room := &entities.Room{
		ID:        uuid.New(),
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.roomStorage.CreateRoom(ctx, room); err != nil {
		return nil, errors.Wrap(err, "failed to create room")
	}
	return room, nil
}

func (s *service) GetRoom(ctx context.Context, roomID uuid.UUID) (*entities.Room, error) {
	room, err := s.roomStorage.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get room")
	}
	return room, nil
}

func (s *service) GetOwnerRooms(ctx context.Context, ownerID uuid.UUID) ([]*entities.Room, error) {
	rooms, err := s.roomStorage.GetRoomsByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get owner's rooms")
	}
	return rooms, nil
}

func (s *service) SearchRooms(ctx context.Context, name string, limit, offset int) ([]*entities.Room, error) {
	rooms, err := s.roomStorage.GetRoomsByName(ctx, name, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search rooms")
	}
	return rooms, nil
}

func (s *service) DeleteRoom(ctx context.Context, roomID, ownerID uuid.UUID) error {
	// Check room ownership.
	room, err := s.roomStorage.GetRoomByID(ctx, roomID)
	if err != nil {
		return errors.Wrap(err, "failed to get room")
	}
	if room.OwnerID != ownerID {
		return entities.ErrRoomDeleteForbidden
	}

	// Delete room.
	if err := s.roomStorage.DeleteRoom(ctx, roomID); err != nil {
		return errors.Wrap(err, "failed to delete room")
	}
	return nil
}

func (s *service) GetAllRooms(ctx context.Context, limit, offset int) ([]*entities.Room, error) {
	rooms, err := s.roomStorage.GetAllRooms(ctx, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all rooms")
	}
	return rooms, nil
}
