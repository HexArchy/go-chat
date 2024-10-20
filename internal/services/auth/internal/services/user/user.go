package user

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Service interface {
	RegisterUser(ctx context.Context, user *entities.User) error
	GetUser(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	CheckPermission(ctx context.Context, userID uuid.UUID, requiredPermission entities.Permission) (bool, error)
	GetUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
}

type service struct {
	userStorage UserStorage
}

func NewService(deps Deps) Service {
	return &service{
		userStorage: deps.UserStorage,
	}
}

func (s *service) RegisterUser(ctx context.Context, user *entities.User) error {
	existingUser, err := s.userStorage.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return errors.New("email already in use")
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return errors.Wrap(err, "failed to hash password")
	}
	user.Password = hashedPassword

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if err := s.userStorage.CreateUser(ctx, user); err != nil {
		return errors.Wrap(err, "failed to register user")
	}
	return nil
}

func (s *service) GetUser(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}
	return user, nil
}

func (s *service) UpdateUser(ctx context.Context, user *entities.User) error {
	existingUser, err := s.userStorage.GetUserByID(ctx, user.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get existing user")
	}

	if user.Email != "" && user.Email != existingUser.Email {
		existingUser.Email = user.Email
	}
	if user.Username != "" && user.Username != existingUser.Username {
		existingUser.Username = user.Username
	}
	if user.Phone != "" && user.Phone != existingUser.Phone {
		existingUser.Phone = user.Phone
	}
	if user.Age != 0 && user.Age != existingUser.Age {
		existingUser.Age = user.Age
	}
	if user.Bio != "" && user.Bio != existingUser.Bio {
		existingUser.Bio = user.Bio
	}
	if len(user.Permissions) > 0 {
		existingUser.Permissions = user.Permissions
	}

	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return errors.Wrap(err, "failed to hash password")
		}
		existingUser.Password = hashedPassword
	}

	if err := s.userStorage.UpdateUser(ctx, existingUser); err != nil {
		return errors.Wrap(err, "failed to update user")
	}
	return nil
}

func (s *service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userStorage.DeleteUser(ctx, userID); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}
	return nil
}

func (s *service) CheckPermission(ctx context.Context, userID uuid.UUID, requiredPermission entities.Permission) (bool, error) {
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return false, errors.Wrap(err, "failed to get user")
	}

	for _, perm := range user.Permissions {
		if perm == requiredPermission {
			return true, nil
		}
	}

	return false, nil
}

func (s *service) GetUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	return s.userStorage.GetUsers(ctx, limit, offset)
}
