package storage

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Storage interface {
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
}

type storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) Storage {
	return &storage{db: db}
}

func (s *storage) CreateUser(ctx context.Context, user *entities.User) error {
	var permissions []*Permission
	for _, perm := range user.Permissions {
		permDTO := &Permission{Name: string(perm)}
		if err := s.db.WithContext(ctx).FirstOrCreate(permDTO, Permission{Name: string(perm)}).Error; err != nil {
			return errors.Wrap(err, "failed to find or create permission")
		}
		permissions = append(permissions, permDTO)
	}

	dto := User{
		ID:          user.ID,
		Email:       user.Email,
		Password:    user.Password,
		Username:    user.Username,
		Phone:       user.Phone,
		Age:         user.Age,
		Bio:         user.Bio,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	if err := s.db.WithContext(ctx).Create(&dto).Error; err != nil {
		return errors.Wrap(err, "failed to create user")
	}
	return nil
}

func (s *storage) GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	var dto User
	if err := s.db.WithContext(ctx).
		Preload("Permissions").
		First(&dto, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrUserNotFound
		}
		return nil, errors.Wrap(err, "failed to get user by ID")
	}
	return dtoToEntity(&dto), nil
}

func (s *storage) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var dto User
	if err := s.db.WithContext(ctx).
		Preload("Permissions").
		First(&dto, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrUserNotFound
		}
		return nil, errors.Wrap(err, "failed to get user by email")
	}
	return dtoToEntity(&dto), nil
}

func (s *storage) UpdateUser(ctx context.Context, user *entities.User) error {
	var dto User
	if err := s.db.WithContext(ctx).
		Preload("Permissions").
		First(&dto, "id = ?", user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.ErrUserNotFound
		}
		return errors.Wrap(err, "failed to find user for update")
	}

	dto.Email = user.Email
	dto.Password = user.Password
	dto.Username = user.Username
	dto.Phone = user.Phone
	dto.Age = user.Age
	dto.Bio = user.Bio
	dto.UpdatedAt = user.UpdatedAt

	// Обновляем permissions
	var permissions []*Permission
	for _, perm := range user.Permissions {
		permDTO := &Permission{Name: string(perm)}
		if err := s.db.WithContext(ctx).FirstOrCreate(permDTO, Permission{Name: string(perm)}).Error; err != nil {
			return errors.Wrap(err, "failed to find or create permission")
		}
		permissions = append(permissions, permDTO)
	}
	if err := s.db.WithContext(ctx).Model(&dto).Association("Permissions").Replace(permissions); err != nil {
		return errors.Wrap(err, "failed to update permissions")
	}

	if err := s.db.WithContext(ctx).Save(&dto).Error; err != nil {
		return errors.Wrap(err, "failed to update user")
	}
	return nil
}

func (s *storage) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Delete(&User{}, "id = ?", userID).Error; err != nil {
		return errors.Wrap(err, "failed to delete user")
	}
	return nil
}

func dtoToEntity(dto *User) *entities.User {
	var permissions []entities.Permission
	for _, permDTO := range dto.Permissions {
		permissions = append(permissions, entities.Permission(permDTO.Name))
	}
	return &entities.User{
		ID:          dto.ID,
		Email:       dto.Email,
		Password:    dto.Password,
		Username:    dto.Username,
		Phone:       dto.Phone,
		Age:         dto.Age,
		Bio:         dto.Bio,
		Permissions: permissions,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}
}

func (s *storage) GetUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := s.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get users")
	}
	return users, nil
}
