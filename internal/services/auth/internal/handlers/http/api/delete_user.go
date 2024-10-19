// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"errors"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/models"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/user_management"
)

func (h *Handler) DeleteUserHandler(params user_management.DeleteUserParams, principal interface{}) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	authUser, ok := principal.(*entities.AuthenticatedUser)
	if !ok {
		return user_management.NewDeleteUserUnauthorized().WithPayload(&models.Error{
			Message: "Invalid authentication credentials",
		})
	}

	if err := h.useCases.DeleteUser.Execute(ctx, params.UserID, authUser.UserID); err != nil {
		return h.handleDeleteUserError(err)
	}

	return user_management.NewDeleteUserNoContent()
}

func (h *Handler) handleDeleteUserError(err error) middleware.Responder {
	switch {
	case errors.Is(err, entities.ErrUserNotFound):
		return user_management.NewDeleteUserNotFound().WithPayload(&models.Error{
			Message: "User not found",
		})
	case errors.Is(err, entities.ErrForbidden):
		return user_management.NewDeleteUserForbidden().WithPayload(&models.Error{
			Message: "Insufficient permissions to delete this user",
		})
	case errors.Is(err, entities.ErrUnauthorized):
		return user_management.NewDeleteUserUnauthorized().WithPayload(&models.Error{
			Message: "Unauthorized access",
		})
	default:
		h.logger.Error("Unexpected error while deleting user", zap.Error(err))
		return user_management.NewDeleteUserInternalServerError().WithPayload(&models.Error{
			Message: "An unexpected error occurred",
		})
	}
}
