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

func (h *Handler) LoginUserHandler(params authentication.LoginUserParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	authenticatedUser, err := h.useCases.LoginUser.Execute(ctx, params.Credentials.Email.String(), params.Credentials.Password.String())
	if err != nil {
		return h.handleLoginError(err)
	}

	return authentication.NewLoginUserOK().WithPayload(transformation.MapAuthenticatedUserToResponse(authenticatedUser))
}

func (h *Handler) handleLoginError(err error) middleware.Responder {
	switch {
	case errors.Is(err, entities.ErrInvalidCredantials):
		return authentication.NewLoginUserUnauthorized().WithPayload(&models.Error{
			Message: "Invalid email or password",
		})
	case errors.Is(err, entities.ErrInvalidInput):
		return authentication.NewLoginUserBadRequest().WithPayload(&models.Error{
			Message: err.Error(),
		})
	default:
		h.logger.Error("Unexpected error during user login", zap.Error(err))
		return authentication.NewLoginUserInternalServerError().WithPayload(&models.Error{
			Message: "An unexpected error occurred",
		})
	}
}
