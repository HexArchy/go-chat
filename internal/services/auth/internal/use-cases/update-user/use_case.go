package updateuser

import (
	"context"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/pkg/errors"
)

type UseCase struct {
	authService AuthService
	authStorage AuthStorage
	rbac        RBAC
}

func New(authService AuthService, authStorage AuthStorage, rbac RBAC) *UseCase {
	return &UseCase{
		authService: authService,
		authStorage: authStorage,
		rbac:        rbac,
	}
}

func (uc *UseCase) Execute(ctx context.Context, input entities.User, updaterID string) (entities.User, error) {
	// Get the updater's authenticated user information.
	updater, err := uc.authStorage.GetUserByID(ctx, updaterID)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "failed to get updater user")
	}

	// Check if the updater has permission to update the user.
	if updater.ID != input.ID && !uc.rbac.CheckPermission(updater.Roles, "update", "user") {
		return entities.User{}, errors.Wrap(entities.ErrForbidden, "user does not have permission to update this user")
	}

	// Get the current user data.
	user, err := uc.authStorage.GetUserByID(ctx, input.ID)
	if err != nil {
		return entities.User{}, errors.Wrap(err, "failed to get user")
	}

	// Update user fields.
	user.Email = input.Email
	user.Name = input.Name
	user.Nickname = input.Nickname
	user.PhoneNumber = input.PhoneNumber
	user.Age = input.Age
	user.Bio = input.Bio

	// Only allow role updates if the updater has admin privileges.
	if uc.rbac.CheckPermission(updater.Roles, "update", "user_roles") {
		user.Roles = input.Roles
	}

	// Validate updated user data.
	if err := user.Validate(); err != nil {
		return entities.User{}, errors.Wrap(entities.ErrInvalidInput, err.Error())
	}

	// Update user in storage.
	if err := uc.authStorage.UpdateUser(ctx, user); err != nil {
		return entities.User{}, errors.Wrap(err, "failed to update user")
	}

	return *user, nil
}
