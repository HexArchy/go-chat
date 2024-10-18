package registeruser

import (
	"context"
	"time"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UseCase struct {
	authService AuthService
	authStorage AuthStorage
}

func New(authService AuthService, authStorage AuthStorage) *UseCase {
	return &UseCase{
		authService: authService,
		authStorage: authStorage,
	}
}

func (uc *UseCase) Execute(ctx context.Context, user entities.User) (entities.AuthenticatedUser, error) {
	// Validate input.
	if err := user.Validate(); err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(entities.ErrInvalidInput, err.Error())
	}

	// Check if user already exists.
	existingUser, err := uc.authStorage.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "failed to check existing user")
	}
	if existingUser != nil {
		return entities.AuthenticatedUser{}, entities.ErrUserAlreadyExists
	}

	// Create new user.
	newUser := &entities.User{
		ID:          uuid.New().String(),
		Email:       user.Email,
		Password:    user.Password,
		Name:        user.Name,
		Nickname:    user.Nickname,
		PhoneNumber: user.PhoneNumber,
		Age:         user.Age,
		Bio:         user.Bio,
		Roles:       []entities.Role{entities.RoleUser},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.authService.RegisterUser(ctx, newUser); err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "failed to register user")
	}

	// Authenticate user and generate tokens.
	authenticatedUser, err := uc.authService.AuthenticateUser(ctx, user.Email, user.Password)
	if err != nil {
		return entities.AuthenticatedUser{}, errors.Wrap(err, "failed to authenticate user")
	}

	return authenticatedUser, nil
}
