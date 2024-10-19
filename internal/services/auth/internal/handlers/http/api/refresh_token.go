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

func (h *Handler) RefreshTokenHandler(params authentication.RefreshTokenParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	if params.RefreshToken == nil || params.RefreshToken.RefreshToken == nil {
		return authentication.NewRefreshTokenBadRequest().WithPayload(&models.Error{
			Message: "Refresh token is required",
		})
	}

	authenticatedUser, err := h.useCases.RefreshToken.Execute(ctx, *params.RefreshToken.RefreshToken)
	if err != nil {
		return h.handleRefreshTokenError(err)
	}

	return authentication.NewRefreshTokenOK().WithPayload(transformation.MapAuthenticatedUserToResponse(authenticatedUser))
}

func (h *Handler) handleRefreshTokenError(err error) middleware.Responder {
	switch {
	case errors.Is(err, entities.ErrInvalidRefreshToken):
		return authentication.NewRefreshTokenUnauthorized().WithPayload(&models.Error{
			Message: "Invalid refresh token",
		})
	case errors.Is(err, entities.ErrInvalidInput):
		return authentication.NewRefreshTokenBadRequest().WithPayload(&models.Error{
			Message: err.Error(),
		})
	default:
		h.logger.Error("Unexpected error during token refresh", zap.Error(err))
		return authentication.NewRefreshTokenInternalServerError().WithPayload(&models.Error{
			Message: "An unexpected error occurred",
		})
	}
}
