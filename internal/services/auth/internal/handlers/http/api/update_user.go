// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"errors"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/models"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/user_management"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/transformation"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
)

func (h *Handler) UpdateUserHandler(params user_management.UpdateUserParams, principal interface{}) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	authUser, ok := principal.(*entities.AuthenticatedUser)
	if !ok {
		return user_management.NewUpdateUserUnauthorized().WithPayload(&models.Error{
			Message: "Invalid authentication credentials",
		})
	}

	updateInput := entities.User{
		ID:          params.UserID,
		Name:        params.User.Name,
		Nickname:    params.User.Nickname,
		PhoneNumber: params.User.PhoneNumber,
		Age:         int(params.User.Age),
		Bio:         params.User.Bio,
	}

	updatedUser, err := h.useCases.UpdateUser.Execute(ctx, updateInput, authUser.UserID)
	if err != nil {
		return h.handleUpdateUserError(err)
	}

	return user_management.NewUpdateUserOK().WithPayload(transformation.MapUserToResponse(updatedUser))
}

func (h *Handler) handleUpdateUserError(err error) middleware.Responder {
	switch {
	case errors.Is(err, entities.ErrUserNotFound):
		return user_management.NewUpdateUserNotFound().WithPayload(&models.Error{
			Message: "User not found",
		})
	case errors.Is(err, entities.ErrInvalidInput):
		return user_management.NewUpdateUserBadRequest().WithPayload(&models.Error{
			Message: err.Error(),
		})
	case errors.Is(err, entities.ErrForbidden):
		return user_management.NewUpdateUserForbidden().WithPayload(&models.Error{
			Message: "Insufficient permissions to update this user",
		})
	default:
		h.logger.Error("Unexpected error during user update", zap.Error(err))
		return user_management.NewUpdateUserInternalServerError().WithPayload(&models.Error{
			Message: "An unexpected error occurred",
		})
	}
}
