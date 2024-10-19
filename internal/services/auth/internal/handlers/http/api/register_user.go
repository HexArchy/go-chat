// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"errors"

	"github.com/HexArch/go-chat/internal/services/auth/internal/entities"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/models"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/authentication"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/transformation"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
)

func (h *Handler) RegisterUserHandler(params authentication.RegisterUserParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	var name string
	if params.User.Name != nil {
		name = *params.User.Name
	}
	user := entities.User{
		Email:       params.User.Email.String(),
		Password:    params.User.Password.String(),
		Name:        name,
		Nickname:    params.User.Nickname,
		PhoneNumber: params.User.PhoneNumber,
		Age:         int(params.User.Age),
		Bio:         params.User.Bio,
	}

	authenticatedUser, err := h.useCases.RegisterUser.Execute(ctx, user)
	if err != nil {
		return h.handleRegisterUserError(err)
	}

	return authentication.NewRegisterUserCreated().WithPayload(transformation.MapAuthenticatedUserToResponse(authenticatedUser))
}

func (h *Handler) handleRegisterUserError(err error) middleware.Responder {
	switch {
	case errors.Is(err, entities.ErrUserAlreadyExists):
		return authentication.NewRegisterUserConflict().WithPayload(&models.Error{
			Message: "User with this email already exists",
		})
	case errors.Is(err, entities.ErrInvalidInput):
		return authentication.NewRegisterUserBadRequest().WithPayload(&models.Error{
			Message: err.Error(),
		})
	default:
		h.logger.Error("Unexpected error during user registration", zap.Error(err))
		return authentication.NewRegisterUserInternalServerError().WithPayload(&models.Error{
			Message: "An unexpected error occurred",
		})
	}
}
