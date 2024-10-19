package deleteuser

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

func (uc *UseCase) Execute(ctx context.Context, userIDToDelete string, deleterID string) error {
	// Get the deleter's authenticated user information.
	deleter, err := uc.authStorage.GetUserByID(ctx, deleterID)
	if err != nil {
		return errors.Wrap(err, "failed to get deleter user")
	}

	// Check if the deleter has permission to delete the user.
	if deleter.ID != userIDToDelete && !uc.rbac.CheckPermission(deleter.Roles, "delete", "user") {
		return errors.Wrap(entities.ErrForbidden, "user does not have permission to delete this user")
	}

	// Get the user to be deleted.
	userToDelete, err := uc.authStorage.GetUserByID(ctx, userIDToDelete)
	if err != nil {
		return errors.Wrap(err, "failed to get user to delete")
	}

	if userToDelete == nil {
		return errors.Wrap(entities.ErrUserNotFound, "user to delete not found")
	}

	// Check if trying to delete an admin (optional, depending on your requirements).
	if uc.rbac.CheckPermission(userToDelete.Roles, "admin", "system") && !uc.rbac.CheckPermission(deleter.Roles, "super_admin", "system") {
		return errors.Wrap(entities.ErrForbidden, "cannot delete an admin user without super admin privileges")
	}

	// Delete user from storage.
	if err := uc.authStorage.DeleteUser(ctx, userIDToDelete); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	// Revoke all refresh tokens for the deleted user.
	if err := uc.authService.RevokeAllRefreshTokens(ctx, userIDToDelete); err != nil {
		return errors.Wrap(err, "failed to revoke refresh tokens")
	}

	return nil
}
