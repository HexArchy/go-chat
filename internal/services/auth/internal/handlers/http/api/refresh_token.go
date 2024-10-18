// This file is safe to edit. Once it exists it will not be overwritten

package api

import (
	"github.com/go-openapi/runtime/middleware"

	apiAuthentication "github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api/restapi/operations/authentication"
)

func (h *Handler) RefreshTokenHandler(params apiAuthentication.RefreshTokenParams) middleware.Responder {
	return middleware.NotImplemented("operation authentication RefreshToken has not yet been implemented")
}
